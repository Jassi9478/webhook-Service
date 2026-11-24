package domain

import (
	"errors"
	"time"
)

var (
	ErrEmptyPayload        = errors.New("event payload cannot be empty")
	ErrEmptyIdempotencyKey = errors.New("idempotency key is required")
)

type DeliveryStatus string

const (
	StatusPending    DeliveryStatus = "PENDING"
	StatusSuccess    DeliveryStatus = "SUCCESS"
	StatusFailed     DeliveryStatus = "FAILED"
	StatusDeadLetter DeliveryStatus = "DLQ"
)

type Event struct {
	ID             string
	TenantID       string
	EventType      string
	Payload        string
	IdempotencyKey string
	CreatedAt      time.Time
}

type DeliveryAttempt struct {
	ID            string
	EventID       string
	TargetURL     string
	Status        DeliveryStatus
	ResponseCode  int
	AttemptNumber int
	CreatedAt     time.Time
}

func NewEvent(id, tenantID, eventType, idempotencyKey, payload string) (*Event, error) {
	if payload == "" {
		return nil, ErrEmptyPayload
	}
	if idempotencyKey == "" {
		return nil, ErrEmptyIdempotencyKey
	}
	return &Event{
		ID:             id,
		TenantID:       tenantID,
		EventType:      eventType,
		Payload:        payload,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now().UTC(),
	}, nil
}
