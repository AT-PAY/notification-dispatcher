package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"notification-dispatcher/internal/api"
	"notification-dispatcher/internal/config"
	"notification-dispatcher/internal/dispatcher"
	"os"
	"os/signal"
	"syscall"
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

	// 1. Create channel to listen from OS
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// 2. Run server in private Goroutine to don't block main flow
	go func() {
		log.Printf("Server starting on port %s...\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// 3. Wait here util receive Ctrl+C
	<-stop
	log.Println("\nShutdown signal received. Starting graceful shutdown...")

	// 4. Setup time wait maximum for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 5. First: Stop request new HTTP
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown error: %v", err)
	}

	// 6. Continue: Close Dispatcher and waiting Worker handle all message
	d.Shutdown()

	log.Println("Server exited gracefully.")
}
