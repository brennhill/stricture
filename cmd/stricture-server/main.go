package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/stricture/stricture/internal/server"
)

func main() {
	cfg := server.LoadConfigFromEnv()
	app, err := server.New(cfg)
	if err != nil {
		log.Fatalf("stricture-server init failed: %v", err)
	}

	log.Printf("stricture-server listening on %s (data_dir=%s)", cfg.Addr, cfg.DataDir)
	if err := app.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("stricture-server exited with error: %v", err)
	}
}
