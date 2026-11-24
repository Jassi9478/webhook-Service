package usecase

import (
	"context"
	"fmt"
	"webhook-service/internal/domain"
)

type IngestEventUseCase struct {
	repo        EventRepository
	idempotency IdempotencyStore
	publisher   QueuePublisher
}

func NewIngestEventUseCase(r EventRepository, i IdempotencyStore, p QueuePublisher) *IngestEventUseCase {
	return &IngestEventUseCase{repo: r, idempotency: i, publisher: p}
}

func (uc *IngestEventUseCase) Execute(ctx context.Context, eventID, tenantID, eventType, key, payload string) (*domain.Event, error) {
	// 1. Check Idempotency Lock
	acquired, err := uc.idempotency.Lock(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("idempotency check failed: %w", err)
	}
	if !acquired {
		return nil, fmt.Errorf("duplicate request with key: %s", key)
	}

	// 2. Build Domain Object
	event, err := domain.NewEvent(eventID, tenantID, eventType, key, payload)
	if err != nil {
		return nil, err
	}

	// 3. Persist Event
	if err := uc.repo.SaveEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	// 4. Publish to Queue asynchronously
	if err := uc.publisher.Publish(ctx, event.ID); err != nil {
		return nil, fmt.Errorf("failed to publish event to queue: %w", err)
	}

	return event, nil
}
