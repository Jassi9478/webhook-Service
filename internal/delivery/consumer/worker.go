package consumer

import (
	"context"
	"log"
	"sync"
	"webhook-service/internal/usecase"
)

type WorkerPool struct {
	numWorkers int
	deliverUC  *usecase.DeliverWebhookUseCase
	queue      usecase.QueueSubscriber
	targetURL  string
	secretKey  string
}

func NewWorkerPool(numWorkers int, uc *usecase.DeliverWebhookUseCase, q usecase.QueueSubscriber, targetURL, secretKey string) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		deliverUC:  uc,
		queue:      q,
		targetURL:  targetURL,
		secretKey:  secretKey,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	jobChan := make(chan string, 100)
	var wg sync.WaitGroup

	// Worker Goroutines
	for i := 1; i <= wp.numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for eventID := range jobChan {
				log.Printf("[Worker %d] Delivering event: %s", workerID, eventID)
				if err := wp.deliverUC.Execute(ctx, eventID, wp.targetURL, wp.secretKey); err != nil {
					log.Printf("[Worker %d] Failed event %s: %v", workerID, eventID, err)
				}
			}
		}(i)
	}

	// Consume from NATS and push to worker channel
	err := wp.queue.Consume(ctx, func(eventID string) error {
		jobChan <- eventID
		return nil
	})
	if err != nil {
		log.Printf("Queue consumer error: %v", err)
	}

	close(jobChan)
	wg.Wait()
}
