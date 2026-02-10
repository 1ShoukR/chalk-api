package events

import (
	"chalk-api/pkg/models"
	"chalk-api/pkg/repositories"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Publisher writes events into the transactional outbox.
type Publisher struct {
	outbox *repositories.OutboxRepository
}

func NewPublisher(outbox *repositories.OutboxRepository) *Publisher {
	return &Publisher{outbox: outbox}
}

func (p *Publisher) Publish(
	ctx context.Context,
	eventType EventType,
	aggregateType string,
	aggregateID string,
	idempotencyKey string,
	payload any,
) error {
	event, err := buildOutboxEvent(eventType, aggregateType, aggregateID, idempotencyKey, payload)
	if err != nil {
		return err
	}
	return p.outbox.Enqueue(ctx, event)
}

func (p *Publisher) PublishInTx(
	ctx context.Context,
	tx *gorm.DB,
	eventType EventType,
	aggregateType string,
	aggregateID string,
	idempotencyKey string,
	payload any,
) error {
	event, err := buildOutboxEvent(eventType, aggregateType, aggregateID, idempotencyKey, payload)
	if err != nil {
		return err
	}
	return p.outbox.EnqueueTx(ctx, tx, event)
}

func buildOutboxEvent(
	eventType EventType,
	aggregateType string,
	aggregateID string,
	idempotencyKey string,
	payload any,
) (*models.OutboxEvent, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal outbox payload: %w", err)
	}

	return &models.OutboxEvent{
		EventType:      string(eventType),
		AggregateType:  aggregateType,
		AggregateID:    aggregateID,
		IdempotencyKey: idempotencyKey,
		Payload:        string(raw),
		Status:         models.OutboxStatusPending,
		AvailableAt:    time.Now().UTC(),
	}, nil
}
