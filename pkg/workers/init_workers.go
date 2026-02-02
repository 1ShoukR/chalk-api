package workers

import (
	"chalk-api/pkg/config"
	"log/slog"
)

// WorkersCollection contains all background workers
type WorkersCollection struct {
	// Add worker fields here as you create them
	// Example:
	// EmailWorker *EmailWorker
	// CleanupWorker *CleanupWorker
}

// InitializeWorkers initializes all background workers
func InitializeWorkers(cfg config.Environment) (*WorkersCollection, error) {
	return &WorkersCollection{
		// Initialize workers here
	}, nil
}

// StartAll starts all background workers
func (w *WorkersCollection) StartAll(cfg config.Environment) {
	slog.Info("Starting all workers...")
	// Start workers here
	// Example:
	// go w.EmailWorker.Start()
}

// StopAll stops all background workers
func (w *WorkersCollection) StopAll() {
	slog.Info("Stopping all workers...")
	// Stop workers here
	// Example:
	// w.EmailWorker.Stop()
}
