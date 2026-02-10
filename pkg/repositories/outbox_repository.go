package repositories

import (
	"chalk-api/pkg/models"
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

// Enqueue inserts an outbox event. Duplicate idempotency keys are treated as success.
func (r *OutboxRepository) Enqueue(ctx context.Context, event *models.OutboxEvent) error {
	db := r.db.WithContext(ctx)
	return r.enqueueWithDB(db, event)
}

// EnqueueTx inserts an outbox event inside an existing transaction.
// Use this with domain writes in the same transaction for reliability.
func (r *OutboxRepository) EnqueueTx(ctx context.Context, tx *gorm.DB, event *models.OutboxEvent) error {
	db := tx.WithContext(ctx)
	return r.enqueueWithDB(db, event)
}

func (r *OutboxRepository) enqueueWithDB(db *gorm.DB, event *models.OutboxEvent) error {
	now := time.Now().UTC()
	if event.Status == "" {
		event.Status = models.OutboxStatusPending
	}
	if event.AvailableAt.IsZero() {
		event.AvailableAt = now
	}

	err := db.Create(event).Error
	if err != nil && isUniqueViolation(err) {
		// Idempotency key already exists: treat as successful publish.
		return nil
	}
	return err
}

// ClaimPending atomically claims pending events ready for processing.
// It uses row locks with SKIP LOCKED so multiple workers can run safely.
func (r *OutboxRepository) ClaimPending(ctx context.Context, limit int) ([]models.OutboxEvent, error) {
	if limit <= 0 {
		limit = 25
	}

	now := time.Now().UTC()
	events := make([]models.OutboxEvent, 0, limit)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND available_at <= ?", models.OutboxStatusPending, now).
			Order("available_at ASC, id ASC").
			Limit(limit).
			Find(&events).Error; err != nil {
			return err
		}

		if len(events) == 0 {
			return nil
		}

		ids := make([]uint, len(events))
		for i := range events {
			ids[i] = events[i].ID
		}

		if err := tx.Model(&models.OutboxEvent{}).
			Where("id IN ?", ids).
			Updates(map[string]any{
				"status":             models.OutboxStatusProcessing,
				"processing_started": now,
				"updated_at":         now,
			}).Error; err != nil {
			return err
		}

		for i := range events {
			events[i].Status = models.OutboxStatusProcessing
			events[i].ProcessingStarted = &now
		}

		return nil
	})

	return events, err
}

func (r *OutboxRepository) MarkProcessed(ctx context.Context, id uint) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":             models.OutboxStatusProcessed,
			"processed_at":       now,
			"processing_started": nil,
			"last_error":         nil,
			"updated_at":         now,
		}).Error
}

func (r *OutboxRepository) MarkRetry(ctx context.Context, id uint, attempts int, lastError string, nextAttemptAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":             models.OutboxStatusPending,
			"attempts":           attempts,
			"last_error":         lastError,
			"available_at":       nextAttemptAt.UTC(),
			"processing_started": nil,
			"updated_at":         time.Now().UTC(),
		}).Error
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id uint, attempts int, lastError string) error {
	return r.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":             models.OutboxStatusFailed,
			"attempts":           attempts,
			"last_error":         lastError,
			"processing_started": nil,
			"updated_at":         time.Now().UTC(),
		}).Error
}

// RequeueStuckProcessing moves stale processing events back to pending.
// Useful when a worker crashes after claiming but before marking status.
func (r *OutboxRepository) RequeueStuckProcessing(ctx context.Context, olderThan time.Duration) (int64, error) {
	if olderThan <= 0 {
		olderThan = 10 * time.Minute
	}

	now := time.Now().UTC()
	cutoff := now.Add(-olderThan)

	result := r.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where("status = ? AND processing_started IS NOT NULL AND processing_started < ?", models.OutboxStatusProcessing, cutoff).
		Updates(map[string]any{
			"status":             models.OutboxStatusPending,
			"available_at":       now,
			"processing_started": nil,
			"updated_at":         now,
		})

	return result.RowsAffected, result.Error
}

func isUniqueViolation(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "duplicate key value violates unique constraint")
}
