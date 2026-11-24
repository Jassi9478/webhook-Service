package usecase

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
	"webhook-service/internal/domain"
)

type DeliverWebhookUseCase struct {
	repo   EventRepository
	sender WebhookSender
}

func NewDeliverWebhookUseCase(r EventRepository, s WebhookSender) *DeliverWebhookUseCase {
	return &DeliverWebhookUseCase{repo: r, sender: s}
}

func (uc *DeliverWebhookUseCase) Execute(ctx context.Context, eventID, targetURL, secretKey string) error {
	event, err := uc.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	maxRetries := 5
	for attempt := 1; attempt <= maxRetries; attempt++ {
		statusCode, err := uc.sender.Send(ctx, targetURL, secretKey, event.Payload)

		status := domain.StatusSuccess
		if err != nil || statusCode < 200 || statusCode >= 300 {
			status = domain.StatusFailed
			if attempt == maxRetries {
				status = domain.StatusDeadLetter
			}
		}

		// Log attempt in Database
		_ = uc.repo.RecordAttempt(ctx, &domain.DeliveryAttempt{
			ID:            fmt.Sprintf("%s-att-%d", event.ID, attempt),
			EventID:       event.ID,
			TargetURL:     targetURL,
			Status:        status,
			ResponseCode:  statusCode,
			AttemptNumber: attempt,
			CreatedAt:     time.Now().UTC(),
		})

		if status == domain.StatusSuccess {
			return nil
		}

		// Exponential backoff with full jitter
		if attempt < maxRetries {
			backoff := calculateJitterBackoff(attempt)
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("failed to deliver webhook after %d attempts", maxRetries)
}

func calculateJitterBackoff(attempt int) time.Duration {
	base := 2.0
	maxDelay := 30 * time.Second
	temp := float64(time.Second) * math.Pow(base, float64(attempt))
	if temp > float64(maxDelay) {
		temp = float64(maxDelay)
	}
	return time.Duration(rand.Float64() * temp)
}
