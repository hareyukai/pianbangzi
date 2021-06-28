package main

import (
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
)

// a chatting room
type room struct {

	// a channel holds incoming messages
	broadcast chan []byte

	// clients want to join
	join chan *client

	// clients want to leave
	leave chan *client

	// clients in the room
	clients map[*client]bool
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

// make a new room
func newRoom() *room {
	return &room{
		broadcast: make(chan []byte),
		join:      make(chan *client),
		leave:     make(chan *client),
		clients:   make(map[*client]bool),
	}
}

// the room is receiving and broadcasting messages
func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// a client join
			r.clients[client] = true
		case client := <-r.leave:
			// a client leave
			delete(r.clients, client)
			close(client.mch)
		case msg := <-r.broadcast:
			// forward message to all clients
			for client := range r.clients {
				select {
				case client.mch <- msg:
					// send the message
				default:
					// failed to send
					delete(r.clients, client)
					close(client.mch)
				}
			}

		}

	}
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

// serve when a client join
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	// generate client id
	emoji := make([]byte, 4)
	for i, v := range []byte{0xF0, 0x9F, 0x98, 0x81} {
		emoji[i] = v
	}
	emoji[3] += byte(rand.Uint32() % 54)

	client := &client{
		id:   string([]byte(emoji)),
		conn: conn,
		mch:  make(chan []byte, messageBufferSize),
		room: r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
