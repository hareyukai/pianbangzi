package main

import "github.com/gorilla/websocket"

// a user
type client struct {

	// the user's id
	id string

	// a websocket connection
	conn *websocket.Conn

	// a channel on which messages are sent
	mch chan []byte

	// the room that the client is chatting in
	room *room
}

// read messages from client
func (c *client) read() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err == nil {
			c.room.broadcast <- []byte(c.id + ": " + string(msg))
		} else {
			break
		}
	}
	c.conn.Close()
}

// write messages to client
func (c *client) write() {
	for msg := range c.mch {
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
	c.conn.Close()
}
