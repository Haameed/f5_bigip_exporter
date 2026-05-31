package main

import (
	"log"
	"net/http"

	"github.com/Haameed/f5_bigip_exporter/internal/config"
	bigIPHTTP "github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/Haameed/f5_bigip_exporter/pkg/probe"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	log.Printf("Starting bigip_exporter version=%s commit=%s buildDate=%s", version, commit, buildDate)

	if err := config.Init(); err != nil {
		log.Fatalf("Initialization error: %+v", err)
	}

	savedConfig := config.GetConfig()
	if err := bigIPHTTP.Configure(savedConfig); err != nil {
		log.Fatalf("%+v", err)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", probe.Handler)
	http.HandleFunc("/health", probe.HealthHandler)

	log.Printf("F5 BIG-IP exporter is running and listening on %q", savedConfig.Listen)
	if err := http.ListenAndServe(savedConfig.Listen, nil); err != nil {
		log.Fatalf("Unable to serve: %v", err)
	}
}
