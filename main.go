package main

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/db"
	"chalk-api/pkg/external"
	"chalk-api/pkg/handlers"
	"chalk-api/pkg/middleware"
	"chalk-api/pkg/repositories"
	"chalk-api/pkg/server"
	"chalk-api/pkg/services"
	"chalk-api/pkg/workers"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Setup Logging
	slog.SetDefault(middleware.SetupLogger(os.Stdout))

	// Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "err", err)
		os.Exit(1)
	}

	// Initialize database (returns GORM DB)
	gormDB, err := db.InitializeDatabase(cfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.CloseDatabase()

	// Run migrations
	err = db.RunMigrations(gormDB)
	if err != nil {
		slog.Error("Problem running migrations", "err", err)
		os.Exit(1)
	}

	// Initialize Repositories
	repositoriesCollection, err := repositories.InitializeRepositories(gormDB)
	if err != nil {
		slog.Error("Failed to initialize repositories", "error", err)
		os.Exit(1)
	}

	// Initialize external integrations
	externalCollection := external.Initialize(cfg)

	// Initialize Services
	servicesCollection, err := services.InitializeServices(repositoriesCollection, cfg)
	if err != nil {
		slog.Error("Failed to initialize services", "err", err)
		os.Exit(1)
	}

	// Initialize Workers (outbox processor, background tasks)
	workersCollection, err := workers.InitializeWorkers(cfg, repositoriesCollection, externalCollection)
	if err != nil {
		slog.Error("Failed to initialize workers", "error", err)
		os.Exit(1)
	}
	workersCollection.StartAll(cfg)
	defer workersCollection.StopAll()

	// Initialize Handlers
	handlersCollection, err := handlers.InitializeHandlers(servicesCollection, repositoriesCollection, cfg)
	if err != nil {
		slog.Error("Failed to initialize handlers", "error", err)
		os.Exit(1)
	}

	// Create and Start Server
	s := server.CreateServer(cfg, gormDB, handlersCollection)

	// Channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		port := fmt.Sprintf(":%d", cfg.Port)
		if err := s.Start(port); err != nil {
			errChan <- err
		}
	}()

	// Wait for signal or server error
	select {
	case <-sigChan:
		slog.Info("Received shutdown signal")
		workersCollection.StopAll()
		if err := s.Shutdown(); err != nil {
			slog.Error("Failed to shutdown server", "error", err)
		}
	case err := <-errChan:
		slog.Error("Server error", "error", err)
		workersCollection.StopAll()
		os.Exit(1)
	}

	slog.Info("Main exiting")
}
