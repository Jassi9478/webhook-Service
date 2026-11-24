package usecase

import (
	"context"
	"webhook-service/internal/domain"
)

type EventRepository interface {
	SaveEvent(ctx context.Context, event *domain.Event) error
	GetEventByID(ctx context.Context, id string) (*domain.Event, error)
	RecordAttempt(ctx context.Context, attempt *domain.DeliveryAttempt) error
}

type IdempotencyStore interface {
	Lock(ctx context.Context, key string) (bool, error)
}

type QueuePublisher interface {
	Publish(ctx context.Context, eventID string) error
}

type QueueSubscriber interface {
	Consume(ctx context.Context, handler func(eventID string) error) error
}

type WebhookSender interface {
	Send(ctx context.Context, targetURL, secretKey string, payload string) (int, error)
}
