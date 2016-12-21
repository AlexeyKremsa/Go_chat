package main

import (
	"log"
	"net/http"

	"github.com/AlexeyKremsa/Go_chat/trace"
	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

type room struct {
	// holds incoming messages that shuold be forwarded to other clients
	forward chan []byte

	// channel for clients wishing to join the room
	join chan *client

	//chanel for clients wishing to leave the room
	leave chan *client

	// all current clients in the room
	clients map[*client]bool

	//tracer will receive trace information of activity in the room
	tracer trace.Tracer
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			//joining
			r.clients[client] = true
			r.tracer.Trace("New client joined")

		case client := <-r.leave:
			//leaving
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left")

		case msg := <-r.forward:
			r.tracer.Trace("Message received: ", string(msg))
			//forward message to all clients
			for client := range r.clients {
				client.send <- msg
			}
		}
	}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}

	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}
