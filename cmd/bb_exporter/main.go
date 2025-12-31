package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/dsegura/bbox-exporter/internal/bbox"
	"github.com/dsegura/bbox-exporter/internal/config"
	"github.com/dsegura/bbox-exporter/internal/exporter"
)

func main() {
	cfgPath := flag.String("config", "appsettings.json", "Path to the exporter configuration file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	client, err := bbox.NewClient(cfg.BBoxAPIURL, cfg.BBoxPassword)
	if err != nil {
		log.Fatalf("init BBox client: %v", err)
	}

	exp := exporter.New(client)
	refreshInterval := time.Duration(cfg.BBoxAPIRefreshTime) * time.Second

	if err := exp.Refresh(context.Background()); err != nil {
		log.Printf("initial refresh failed: %v", err)
	}

	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		for range ticker.C {
			if err := exp.Refresh(context.Background()); err != nil {
				log.Printf("refresh failed: %v", err)
			}
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	addr := fmt.Sprintf(":%d", cfg.MetricsServerListeningPort)
	log.Printf("serving metrics at %s/metrics", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("metrics server stopped: %v", err)
	}
}
