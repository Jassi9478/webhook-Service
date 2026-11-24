package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"webhook-service/internal/adapter/nats"
	"webhook-service/internal/adapter/postgres"
	"webhook-service/internal/adapter/redis"
	deliveryHTTP "webhook-service/internal/delivery/http"
	"webhook-service/internal/usecase"
)

func main() {
	// Replace root with postgres
	repo, err := postgres.NewEventRepository("postgres://postgres:secret@localhost:5433/webhooks?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}

	idempotency := redis.NewIdempotencyStore("localhost:6379")

	queue, err := nats.NewQueueClient("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	ingestUC := usecase.NewIngestEventUseCase(repo, idempotency, queue)
	handler := deliveryHTTP.NewIngestHandler(ingestUC)

	server := &http.Server{Addr: ":8080", Handler: handler}

	go func() {
		log.Println("Ingestion API server starting on :8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	log.Println("API Server gracefully stopped.")
}
