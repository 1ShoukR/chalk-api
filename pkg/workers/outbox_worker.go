package workers

import (
	"chalk-api/pkg/events"
	"chalk-api/pkg/models"
	"chalk-api/pkg/repositories"
	"context"
	"log/slog"
	"sync"
	"time"
)

type OutboxWorkerConfig struct {
	PollInterval time.Duration
	BatchSize    int
	MaxAttempts  int
	StuckAfter   time.Duration
}

type OutboxWorker struct {
	repo       *repositories.OutboxRepository
	dispatcher *events.Dispatcher
	config     OutboxWorkerConfig

	stopCh    chan struct{}
	doneCh    chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
}

func NewOutboxWorker(
	repo *repositories.OutboxRepository,
	dispatcher *events.Dispatcher,
	config OutboxWorkerConfig,
) *OutboxWorker {
	if config.PollInterval <= 0 {
		config.PollInterval = 2 * time.Second
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 25
	}
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 8
	}
	if config.StuckAfter <= 0 {
		config.StuckAfter = 10 * time.Minute
	}

	return &OutboxWorker{
		repo:       repo,
		dispatcher: dispatcher,
		config:     config,
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
}

func (w *OutboxWorker) Start() {
	w.startOnce.Do(func() {
		go w.loop()
		slog.Info("Outbox worker started",
			"poll_interval", w.config.PollInterval.String(),
			"batch_size", w.config.BatchSize,
			"max_attempts", w.config.MaxAttempts,
		)
	})
}

func (w *OutboxWorker) Stop() {
	w.stopOnce.Do(func() {
		close(w.stopCh)
		<-w.doneCh
		slog.Info("Outbox worker stopped")
	})
}

func (w *OutboxWorker) loop() {
	defer close(w.doneCh)

	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	// Run immediately on startup.
	w.runCycle()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.runCycle()
		}
	}
}

func (w *OutboxWorker) runCycle() {
	ctx := context.Background()

	if recovered, err := w.repo.RequeueStuckProcessing(ctx, w.config.StuckAfter); err != nil {
		slog.Error("Outbox worker failed to requeue stale events", "error", err)
	} else if recovered > 0 {
		slog.Warn("Outbox worker requeued stale events", "count", recovered)
	}

	eventsToProcess, err := w.repo.ClaimPending(ctx, w.config.BatchSize)
	if err != nil {
		slog.Error("Outbox worker failed to claim events", "error", err)
		return
	}
	if len(eventsToProcess) == 0 {
		return
	}

	for i := range eventsToProcess {
		w.processEvent(ctx, eventsToProcess[i])
	}
}

func (w *OutboxWorker) processEvent(ctx context.Context, eventRecord models.OutboxEvent) {
	err := w.dispatcher.Dispatch(ctx, eventRecord)
	if err == nil {
		if markErr := w.repo.MarkProcessed(ctx, eventRecord.ID); markErr != nil {
			slog.Error("Outbox worker failed to mark event processed", "event_id", eventRecord.ID, "error", markErr)
		}
		return
	}

	attempts := eventRecord.Attempts + 1
	errorMessage := truncateError(err.Error(), 2000)

	if events.IsPermanent(err) || attempts >= w.config.MaxAttempts {
		if failErr := w.repo.MarkFailed(ctx, eventRecord.ID, attempts, errorMessage); failErr != nil {
			slog.Error("Outbox worker failed to mark event failed", "event_id", eventRecord.ID, "error", failErr)
			return
		}

		slog.Error("Outbox event permanently failed",
			"event_id", eventRecord.ID,
			"event_type", eventRecord.EventType,
			"attempts", attempts,
			"error", errorMessage,
		)
		return
	}

	retryAt := time.Now().UTC().Add(backoffForAttempt(attempts))
	if retryErr := w.repo.MarkRetry(ctx, eventRecord.ID, attempts, errorMessage, retryAt); retryErr != nil {
		slog.Error("Outbox worker failed to schedule retry", "event_id", eventRecord.ID, "error", retryErr)
		return
	}

	slog.Warn("Outbox event scheduled for retry",
		"event_id", eventRecord.ID,
		"event_type", eventRecord.EventType,
		"attempts", attempts,
		"retry_at", retryAt,
		"error", errorMessage,
	)
}

// backoffForAttempt uses exponential backoff with a cap.
func backoffForAttempt(attempt int) time.Duration {
	if attempt <= 1 {
		return 15 * time.Second
	}

	delay := 15 * time.Second
	maxDelay := 15 * time.Minute
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay >= maxDelay {
			return maxDelay
		}
	}
	return delay
}

func truncateError(message string, maxLen int) string {
	if len(message) <= maxLen {
		return message
	}
	return message[:maxLen]
}
