package radarserver

import (
	"context"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/wdantuma/signalk-radar/radar"
	"github.com/wdantuma/signalk-radar/stream"
)

var Version = "0.0.1" // overwritten with VERSION DEF during build

const (
	SERVER_NAME string = "signalk-radar"
)

type radarServer struct {
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
	case *radar.RadarMessage:
		return v
	default:
		return &radar.RadarMessage{}
	}
}

func (server *radarServer) SetupServer(ctx context.Context, hostname string, router *mux.Router) *mux.Router {
	// var err error
	if router == nil {
		router = mux.NewRouter()
	}

	signalk := router.PathPrefix("/signalk").Subrouter()
	// signalk.HandleFunc("", server.hello)
	streamHandler := stream.NewStreamHandler(server)
	// vesselHandler := vessel.NewVesselHandler(server)
	// chartsHandler := charts.NewChartsHandler(server.chartsPath)
	signalk.PathPrefix("/v1/stream").Handler(streamHandler)
	// signalk.PathPrefix("/v1/api/vessels").Handler(vesselHandler)
	// signalk.PathPrefix("/v2/api/resources/charts").Handler(chartsHandler)
	signalk.HandleFunc("/v1/api/snapshot", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})

	// server.converter, err = converter.NewCanToSignalk(server)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// canSource := server.sourcehub.Start()
	// converted := server.converter.Convert(canSource)
	// valueStore := store.NewMemoryStore()
	// server.store = valueStore
	// stored := valueStore.Store(converted)

	go func() {
		for {
			cases := make([]reflect.SelectCase, len(server.radars))
			for i, ch := range server.radars {
				cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch.Source())}
			}
			selected, value, ok := reflect.Select(cases)
			if ok {
				message := RadarMessage(value)
				radar := uint32(selected)
				message.Radar = &radar
				streamHandler.BroadcastDelta <- message

			} else {
				break
			}
		}
		//close(sh.output)
	}()

	return router
}
