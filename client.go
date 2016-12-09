package main

import (
	"github.com/gorilla/websocket"
)

type client struct {
	socket *websocket.Conn

	send chan []byte
	room *room
}

func (c *client) read {
	defer c.socket.Clost()

}
