package models

import "time"

const (
	OutboxStatusPending    = "pending"
	OutboxStatusProcessing = "processing"
	OutboxStatusProcessed  = "processed"
	OutboxStatusFailed     = "failed"
)

// OutboxEvent stores domain events for reliable async side-effects.
// Services should write domain changes + outbox row in the same DB transaction.
type OutboxEvent struct {
	ID uint `gorm:"primaryKey" json:"id"`

	EventType      string `gorm:"not null;index:idx_outbox_type_status,priority:1" json:"event_type"`
	AggregateType  string `gorm:"not null;index" json:"aggregate_type"` // "workout", "message", "session"
	AggregateID    string `gorm:"not null;index" json:"aggregate_id"`   // string to support both numeric and external IDs
	IdempotencyKey string `gorm:"uniqueIndex;not null" json:"idempotency_key"`

	// Payload is JSON encoded event data.
	Payload string `gorm:"type:jsonb;not null" json:"payload"`

	Status            string     `gorm:"not null;default:'pending';index:idx_outbox_status_available,priority:1;index:idx_outbox_type_status,priority:2" json:"status"`
	Attempts          int        `gorm:"not null;default:0" json:"attempts"` // failed attempts count
	AvailableAt       time.Time  `gorm:"not null;index:idx_outbox_status_available,priority:2" json:"available_at"`
	ProcessingStarted *time.Time `gorm:"index" json:"processing_started"`
	ProcessedAt       *time.Time `json:"processed_at"`
	LastError         *string    `gorm:"type:text" json:"last_error"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (OutboxEvent) TableName() string {
	return "outbox_events"
}
