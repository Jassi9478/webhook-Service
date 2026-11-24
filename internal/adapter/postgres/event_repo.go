package postgres

import (
	"context"
	"database/sql"
	"webhook-service/internal/domain"

	_ "github.com/lib/pq"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(connStr string) (*EventRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &EventRepository{db: db}, nil
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	query := `INSERT INTO events (id, tenant_id, event_type, payload, idempotency_key, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, event.ID, event.TenantID, event.EventType, event.Payload, event.IdempotencyKey, event.CreatedAt)
	return err
}

func (r *EventRepository) GetEventByID(ctx context.Context, id string) (*domain.Event, error) {
	query := `SELECT id, tenant_id, event_type, payload, idempotency_key, created_at FROM events WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var e domain.Event
	if err := row.Scan(&e.ID, &e.TenantID, &e.EventType, &e.Payload, &e.IdempotencyKey, &e.CreatedAt); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EventRepository) RecordAttempt(ctx context.Context, attempt *domain.DeliveryAttempt) error {
	query := `INSERT INTO delivery_attempts (id, event_id, target_url, status, response_code, attempt_number, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query, attempt.ID, attempt.EventID, attempt.TargetURL, attempt.Status, attempt.ResponseCode, attempt.AttemptNumber, attempt.CreatedAt)
	return err
}
