package stream

import (
	"fmt"

	"github.com/wdantuma/signalk-radar/radar"
	"google.golang.org/protobuf/proto"
)

type hub struct {
	// Registered clients.
	clients map[*client]bool

	// Inbound messages from the clients.
	Broadcast chan *radar.RadarMessage

	// Register requests from the clients.
	register chan *client

	// Unregister requests from clients.
	unregister chan *client
}

func NewHub() *hub {
	hub := &hub{
		Broadcast:  make(chan *radar.RadarMessage),
		register:   make(chan *client),
		unregister: make(chan *client),
		clients:    make(map[*client]bool),
	}
	hub.run()
	return hub
}

func (h *hub) run() {
	go func() {
		for {
			select {
			case client := <-h.register:
				fmt.Println("Register")
				h.clients[client] = true
			case client := <-h.unregister:
				if _, ok := h.clients[client]; ok {
					fmt.Println("Un register")
					delete(h.clients, client)
					close(client.send)
				}
			case message := <-h.Broadcast:
				bytes, err := proto.Marshal(message)
				if err == nil && len(bytes) > 0 {
					for client := range h.clients {
						select {
						case client.send <- bytes:
						default:
							fmt.Println("Send error")
							delete(h.clients, client)
							close(client.send)
						}
					}
				}
			}
		}
	}()
}
