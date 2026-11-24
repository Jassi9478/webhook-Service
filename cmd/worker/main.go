package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"webhook-service/internal/adapter/httpclient"
	"webhook-service/internal/adapter/nats"
	"webhook-service/internal/adapter/postgres"
	"webhook-service/internal/delivery/consumer"
	"webhook-service/internal/usecase"
)

func main() {
	// Replace root with postgres
	repo, err := postgres.NewEventRepository("postgres://postgres:secret@localhost:5433/webhooks?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}

	queue, err := nats.NewQueueClient("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	sender := httpclient.NewHTTPWebhookSender()
	deliverUC := usecase.NewDeliverWebhookUseCase(repo, sender)

	// Target endpoint to simulate webhook delivery recipient
	targetWebhookURL := "https://httpbin.org/post"
	secretKey := "super-secret-client-key"

	workerPool := consumer.NewWorkerPool(10, deliverUC, queue, targetWebhookURL, secretKey)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		log.Println("Webhook Worker Pool started (10 Workers)...")
		workerPool.Start(ctx)
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker pool...")
	cancel()
}
