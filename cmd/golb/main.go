package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Raphasha27/golb/internal/balancer"
	"github.com/Raphasha27/golb/internal/config"
	"github.com/Raphasha27/golb/internal/health"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	lb := balancer.New(cfg)
	hc := health.NewChecker(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hc.Start(ctx)
	defer hc.Stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/", lb.ServeHTTP)
	mux.HandleFunc("/metrics", lb.MetricsHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: mux,
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("shutting down...")
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)
	}()

	log.Printf("golb starting on :%d", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
