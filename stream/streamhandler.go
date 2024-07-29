package stream

import (
	"log"
	"net/http"
	"time"

	"github.com/wdantuma/signalk-radar/radar"
	"github.com/wdantuma/signalk-radar/radarserver/state"
)

type streamHandler struct {
	state          state.ServerState
	BroadcastDelta chan *radar.RadarMessage
	hub            *hub
}

func NewStreamHandler(s state.ServerState) *streamHandler {
	hub := NewHub()
	return &streamHandler{state: s, hub: hub, BroadcastDelta: hub.Broadcast}
}

// serveWs handles websocket requests from the peer.
func (s *streamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &client{hub: s.hub, conn: conn, send: make(chan []byte, 1024)}
	//format.Json(contextFilter.Filter(client.sendDelta), client.send)
	time.Sleep(1 * time.Second)
	client.hub.register <- client

	//client.send <- s.helloMessage()

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
