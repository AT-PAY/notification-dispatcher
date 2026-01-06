package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"notification-dispatcher/internal/api"
	"notification-dispatcher/internal/config"
	"notification-dispatcher/internal/dispatcher"
	"time"
)

func main() {
	cfg := config.LoadConfig()

	d := dispatcher.NewDispatcher(cfg.DefaultChanelCapacity)

	d.StartWorkerPool(cfg.DefaultNumberWorkers)

	h := api.NewHandle(d, cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/send", h.SendNotificationHandle)
	mux.HandleFunc("/ws", h.WSHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server starting on port %s...", cfg.Port)

	// Listener
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Could not listen on %s: %v\n", cfg.Port, err)
	}
}
