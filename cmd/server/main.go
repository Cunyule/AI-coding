package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Cunyule/AI-coding/internal/api"
	"github.com/Cunyule/AI-coding/internal/engine"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		healthcheck(os.Args[2:])
		return
	}
	addr := flag.String("addr", envOr("SERVER_ADDR", ":8080"), "HTTP listen address")
	rulesPath := flag.String("rules", envOr("RULES_PATH", "rules/fingerprints.json"), "fingerprint rules file")
	flag.Parse()

	fingerprinter, err := engine.Load(*rulesPath)
	if err != nil {
		slog.Error("failed to initialize rule engine", "error", err)
		os.Exit(1)
	}
	server := &http.Server{
		Addr:              *addr,
		Handler:           api.Handler(fingerprinter),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		slog.Info("fingerprint server listening", "addr", *addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}

func healthcheck(args []string) {
	flags := flag.NewFlagSet("healthcheck", flag.ExitOnError)
	url := flags.String("url", "http://127.0.0.1:8080/health", "health endpoint")
	_ = flags.Parse(args)
	client := http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(*url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "unhealthy status: %s\n", response.Status)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
