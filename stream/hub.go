package stream

import (
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
				h.clients[client] = true
			case client := <-h.unregister:
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
				}
			case message := <-h.Broadcast:
				bytes, err := proto.Marshal(message)
				if err == nil {
					for client := range h.clients {
						select {
						case client.send <- bytes:
						default:
							close(client.send)
							delete(h.clients, client)
						}
					}
				}
			}
		}
	}()
}
