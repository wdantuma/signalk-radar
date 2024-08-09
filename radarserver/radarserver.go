package radarserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/wdantuma/signalk-radar/radar"
	"github.com/wdantuma/signalk-radar/stream"
)

var Version = "0.0.1" // overwritten with VERSION DEF during build

const (
	SERVER_NAME string = "signalk-radar"
)

type radarServer struct {
	baseUrl string
	radars  map[string]radar.RadarSource
	name    string
	version string
	debug   bool
}

func NewRadarServer() *radarServer {
	return &radarServer{radars: make(map[string]radar.RadarSource)}
}

func (s *radarServer) GetName() string {
	return s.name
}

func (s *radarServer) GetVersion() string {
	return s.version
}

func (s *radarServer) GetDebug() bool {
	return s.debug
}

func (s *radarServer) SetDebug(debug bool) {
	s.debug = debug
}

func (s *radarServer) AddRadar(radar radar.RadarSource) {
	id := fmt.Sprintf("radar-%d", len(s.radars))
	s.radars[id] = radar
}

func (s *radarServer) GetRadar(id string) (radar.RadarSource, bool) {
	radar := s.radars[id]
	if radar == nil {
		return nil, false
	}
	return radar, true
}

func RadarMessage(value interface{}) *radar.RadarMessage {
	switch v := value.(type) {
	case radar.RadarMessage:
		return &v
	default:
		return &radar.RadarMessage{}
	}
}

func (server *radarServer) MarshallToJSON(req *http.Request) ([]byte, error) {
	var result map[string]radar.Radar = make(map[string]radar.Radar)
	for id, r := range server.radars {
		streamUrl := url.URL{Scheme: "http", Host: req.Host, Path: fmt.Sprintf("/v1/api/stream/%s", id)}
		result[id] = radar.Radar{Id: id, Name: r.Name(), Spokes: r.Spokes(), MaxSpokeLen: r.MaxSpokeLen(), StreamUrl: streamUrl.String()}
	}
	return json.Marshal(result)
}

func (server *radarServer) list(w http.ResponseWriter, req *http.Request) {
	bytes, _ := server.MarshallToJSON(req)
	w.Write(bytes)
}

func (server *radarServer) SetupServer(ctx context.Context, hostname string, router *mux.Router) *mux.Router {
	if router == nil {
		router = mux.NewRouter()
	}

	radar := router.PathPrefix("/v1/api").Subrouter()
	streamHandler := stream.NewStreamHandler(server)
	radar.HandleFunc("/radars", server.list)
	radar.PathPrefix("/stream/{radarId}").Handler(streamHandler)
	go func() {
		for {
			value := <-server.radars["radar-0"].Source()
			// cases := make([]reflect.SelectCase, len(server.radars))
			// for i, ch := range server.radars {
			// 	cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch.Source())}
			// }
			// selected, value, ok := reflect.Select(cases)
			// if ok {
			message := RadarMessage(value)
			radar := uint32(0) //uint32(selected)
			message.Radar = radar
			streamHandler.BroadcastDelta <- value

			// } else {
			// 	break
			// }
		}
		//close(sh.output)
	}()

	return router
}
