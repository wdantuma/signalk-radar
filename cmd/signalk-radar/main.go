package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/wdantuma/signalk-radar/radar"
	"github.com/wdantuma/signalk-radar/radar/garminxhd"
	"github.com/wdantuma/signalk-radar/radarserver"
	"github.com/wdantuma/signalk-radar/source/pcapsource"
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
	serveWebApps := flag.Bool("webapps", true, "Serve webapps")
	version := flag.Bool("version", false, "Show version")
	port := flag.Int("port", listenPort, "Listen port")
	debug := flag.Bool("debug", false, "Enable debugging")
	staticPath := flag.String("webapp-path", "./webapps", "Path to webapps")
	var fileSources arrayFlag
	flag.Var(&fileSources, "file-source", "Path to pcap file")
	var radars arrayFlag
	flag.Var(&radars, "type", "Radar type")

	flag.Parse()

	if len(fileSources) != len(radars) {
		fmt.Printf("Number of sources and types must be equal\n")
		return
	}

	if len(fileSources) == 0 {
		fmt.Printf("At least one source and type must be given\n")
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
		radarServer.SetDebug(true)
		router.Use(loggingMiddleware)
	}

	if *version {
		fmt.Printf("%s version : %s\n", radarServer.GetName(), radarServer.GetVersion())
		return
	}

	if len(fileSources) > 0 {
		for index, fs := range fileSources {
			source, err := pcapsource.NewPcapSource(fs, true)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
			var radar radar.RadarSource
			switch radars[index] {
			case "garminxhd":
				reportFarmeSource := source.CreateFrameSource("garminReport", 50100)
				dataFrameSource := source.CreateFrameSource("garminData", 50102)
				radar = garminxhd.NewGarminXhd(reportFarmeSource, dataFrameSource)
			default:
				fmt.Printf("Radar type %s not supported\n", radars[index])
				return
			}
			radarServer.AddRadar(radar)
			source.Start()
		}
	}

	radarServer.SetupServer(ctx, "", router)

	if *serveWebApps {
		fmt.Printf("Serving webapps from %s\n", *staticPath)
		// setup static file server at /@signalk
		fs := http.FileServer(http.Dir(*staticPath))
		router.PathPrefix("/").Handler(fs)
	}

	// start listening
	fmt.Printf("Listening on :%d...\n", listenPort)
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
