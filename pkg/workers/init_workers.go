package workers

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/events"
	"chalk-api/pkg/external"
	"chalk-api/pkg/repositories"
	"log/slog"
	"time"
)

// WorkersCollection contains all background workers
type WorkersCollection struct {
	Outbox *OutboxWorker
}

// InitializeWorkers initializes all background workers
func InitializeWorkers(
	cfg config.Environment,
	repos *repositories.RepositoriesCollection,
	integrations *external.Collection,
) (*WorkersCollection, error) {
	dispatcher := events.NewDispatcher()
	if err := events.RegisterDefaultHandlers(dispatcher, integrations); err != nil {
		return nil, err
	}

	outboxWorker := NewOutboxWorker(repos.Outbox, dispatcher, OutboxWorkerConfig{
		PollInterval: time.Duration(cfg.OutboxPollIntervalSeconds) * time.Second,
		BatchSize:    cfg.OutboxBatchSize,
		MaxAttempts:  cfg.OutboxMaxAttempts,
		StuckAfter:   time.Duration(cfg.OutboxStuckThresholdSeconds) * time.Second,
	})

	return &WorkersCollection{
		Outbox: outboxWorker,
	}, nil
}

// StartAll starts all background workers
func (w *WorkersCollection) StartAll(cfg config.Environment) {
	slog.Info("Starting all workers...")
	if w.Outbox != nil {
		w.Outbox.Start()
	}
}

// StopAll stops all background workers
func (w *WorkersCollection) StopAll() {
	slog.Info("Stopping all workers...")
	if w.Outbox != nil {
		w.Outbox.Stop()
	}
}
