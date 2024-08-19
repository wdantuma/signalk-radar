package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/wdantuma/signalk-radar/radar-server/radar"
	"github.com/wdantuma/signalk-radar/radar-server/radar/garminxhd"
	"github.com/wdantuma/signalk-radar/radar-server/radar/navico"
	"github.com/wdantuma/signalk-radar/radar-server/radarserver"
	"github.com/wdantuma/signalk-radar/radar-server/source"
	"github.com/wdantuma/signalk-radar/radar-server/source/pcapsource"
	"github.com/wdantuma/signalk-radar/radar-server/source/udpsource"
)

type arrayFlag []string

func (s *arrayFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *arrayFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func main() {

	ctx := context.Background()
	cfg := tls.Config{}

	var listenPort int = 3001

	enableTls := flag.Bool("tls", false, "Enable tls")
	tlsCertFile := flag.String("tlscert", "", "Tls certificate file")
	tlsKeyFile := flag.String("tlskey", "", "Tls key file")
	serveWebApps := flag.Bool("webapps", false, "Serve webapps")
	version := flag.Bool("version", false, "Show version")
	port := flag.Int("port", listenPort, "Listen port")
	debug := flag.Bool("debug", false, "Enable debugging")
	staticPath := flag.String("webapp-path", "./webapps", "Path to webapps")
	pcapSource := flag.String("pcap-source", "", "Path to pcap file")
	udpSource := flag.Bool("udp-source", false, "Use UDP as source")
	radarType := flag.String("type", "", "Radar type")
	flag.Parse()

	if len(*pcapSource) == 0 && !*udpSource {
		fmt.Printf("A source must be given\n")
		return
	}

	if len(*pcapSource) > 0 && *udpSource {
		fmt.Printf("Only one source may be given\n")
		return
	}

	if len(*radarType) == 0 {
		fmt.Printf("A Radar type must be given\n")
		return
	}

	listenPort = *port

	if *tlsCertFile != "" && *tlsKeyFile != "" && *enableTls {
		cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
		if err != nil {
			log.Fatal(err)
		}

		cfg.Certificates = append(cfg.Certificates, cert)
	}

	router := mux.NewRouter()
	router.Use((loggingMiddleware))
	router.Use(handlers.CORS(
		handlers.AllowCredentials(),
		handlers.AllowedHeaders([]string{"authorization", "content-type", "dpop"}),
		handlers.AllowedOriginValidator(func(_ string) bool {
			return true
		}),
	))
	radarServer := radarserver.NewRadarServer()
	if *debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		radarServer.SetDebug(true)
		router.Use(loggingMiddleware)
	}

	if *version {
		fmt.Printf("%s version : %s\n", radarServer.GetName(), radarServer.GetVersion())
		return
	}

	var source source.FrameSourceFactory
	var err error

	if *udpSource {
		source, err = udpsource.NewUdpSource()
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
	}

	if len(*pcapSource) > 0 {
		source, err = pcapsource.NewPcapSource(*pcapSource, true)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
	}

	var radar radar.RadarSource
	switch *radarType {
	case "garminxhd":
		radar = garminxhd.NewGarminXhd(source)
	case "navico":
		radar = navico.NewNavico(source)
	default:
		fmt.Printf("Radar type %s not supported\n", *radarType)
		return
	}
	radarServer.AddRadar(radar)
	source.Start()

	radarServer.SetupServer(ctx, "", router)

	if *serveWebApps {
		fmt.Printf("Serving webapps from %s\n", *staticPath)
		// setup static file server at /@signalk
		fs := http.FileServer(http.Dir(*staticPath))
		router.PathPrefix("/").Handler(fs)
	}

	// start listening
	fmt.Printf("radar-server started, Listening on :%d , using %s\n", listenPort, source.Label())
	server := http.Server{Addr: fmt.Sprintf(":%d", listenPort), Handler: router, TLSConfig: &cfg}
	if *enableTls {
		err := server.ListenAndServeTLS("", "")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}

	<-ctx.Done()
}
