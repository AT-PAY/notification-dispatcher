package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"notification-dispatcher/database"
	"notification-dispatcher/internal/api"
	"notification-dispatcher/internal/config"
	"notification-dispatcher/internal/consumer"
	"notification-dispatcher/internal/dispatcher"
	"notification-dispatcher/internal/middleware"
	"notification-dispatcher/internal/persistence"
	"notification-dispatcher/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.InitDB(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer db.Close()

	// Initialize repository
	notificationDB := persistence.NewNotificationRepository(db)

	// Initialize template renderer
	renderer := service.NewTemplateEngine()

	// Initialize notification service
	notificationService := service.NewNotificationService(notificationDB, renderer)

	// Initialize dispatcher
	d := dispatcher.NewDispatcher(cfg.DefaultChanelCapacity, cfg.RedisUrl, notificationService)

	d.StartRedisSubscriber()
	d.StartWorkerPool(cfg.DefaultNumberWorkers)

	// Initialize Kafka consumer with all required dependencies
	kafkaConsumer := consumer.NewKafkaConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaTopic,
		cfg.KafkaConsumerGroup,
		d,
		notificationService,
	)

	ctxConsumer, cancelConsumer := context.WithCancel(context.Background())
	go kafkaConsumer.Start(ctxConsumer)

	h := api.NewHandle(d, cfg, notificationService)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/send", h.SendNotificationHandle)
	mux.HandleFunc("/ws", h.WSHandler)

	// config cors
	handler := middleware.CorsMiddleware(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 1. Create channel to listen from OS
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// 2. Run server in private Goroutine to don't block main flow
	go func() {
		log.Printf("🚀 Server starting on port %s...\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	log.Println("✅ All services initialized successfully")

	// 3. Wait here until receive Ctrl+C
	<-stop
	log.Println("\n🛑 Shutdown signal received. Starting graceful shutdown...")

	cancelConsumer()
	kafkaConsumer.Close()

	// 4. Setup time wait maximum for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 5. First: Stop accepting new HTTP requests
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ HTTP server shutdown error: %v", err)
	}

	// 6. Continue: Close Dispatcher and wait for workers to handle all messages
	d.Shutdown()

	log.Println("✅ Server exited gracefully.")
}
