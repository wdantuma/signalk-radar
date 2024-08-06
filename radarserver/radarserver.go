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
	radars  []radar.RadarSource
	name    string
	version string
	debug   bool
}

func NewRadarServer() *radarServer {
	return &radarServer{}
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
	s.radars = append(s.radars, radar)
}

func (s *radarServer) GetRadars() []radar.RadarSource {
	return s.radars
}

func (s *radarServer) GetRadar(index int) (radar.RadarSource, bool) {
	if index < 0 || index > len(s.radars)-1 {
		return nil, false
	}
	return s.radars[index], true
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
	var result []radar.Radar = make([]radar.Radar, 0)
	for index, r := range server.radars {
		streamUrl := url.URL{Scheme: "http", Host: req.Host, Path: fmt.Sprintf("/radar/v1/stream/%d", index)}
		result = append(result,
			radar.Radar{Label: r.Label(), Spokes: r.Spokes(), MaxSpokeLen: r.MaxSpokeLen(), StreamUrl: streamUrl.String()},
		)
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

	radar := router.PathPrefix("/radar").Subrouter()
	streamHandler := stream.NewStreamHandler(server)
	radar.HandleFunc("/v1/radars", server.list)
	radar.PathPrefix("/v1/stream/{id}").Handler(streamHandler)
	go func() {
		for {
			value := <-server.radars[0].Source()
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
