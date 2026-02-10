package events

import (
	"chalk-api/pkg/models"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type Handler interface {
	Handle(ctx context.Context, event models.OutboxEvent) error
}

type HandlerFunc func(ctx context.Context, event models.OutboxEvent) error

func (h HandlerFunc) Handle(ctx context.Context, event models.OutboxEvent) error {
	return h(ctx, event)
}

// NonRetryableError marks errors that should not be retried.
type NonRetryableError struct {
	Err error
}

func (e NonRetryableError) Error() string {
	return e.Err.Error()
}

func (e NonRetryableError) Unwrap() error {
	return e.Err
}

func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return NonRetryableError{Err: err}
}

func IsPermanent(err error) bool {
	var target NonRetryableError
	return errors.As(err, &target)
}

// Dispatcher routes outbox events to handlers by event_type.
// Keep one handler per event_type to avoid duplicate side-effects during retries.
type Dispatcher struct {
	handlers map[string]Handler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]Handler),
	}
}

func (d *Dispatcher) Register(eventType EventType, handler Handler) error {
	key := string(eventType)
	if _, exists := d.handlers[key]; exists {
		return fmt.Errorf("handler already registered for event type %s", key)
	}
	d.handlers[key] = handler
	return nil
}

func (d *Dispatcher) Dispatch(ctx context.Context, event models.OutboxEvent) error {
	handler, ok := d.handlers[event.EventType]
	if !ok {
		slog.Debug("No handler registered for outbox event", "event_type", event.EventType, "event_id", event.ID)
		return nil
	}

	return handler.Handle(ctx, event)
}
