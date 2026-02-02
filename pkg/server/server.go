package server

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/handlers"
	"chalk-api/pkg/routes"
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	Config     *config.Environment
	Router     *gin.Engine
	DB         *gorm.DB
	httpServer *http.Server
}

// CreateServer initializes and returns a configured server instance
func CreateServer(cfg config.Environment, db *gorm.DB, handlers *handlers.HandlersCollection) *Server {
	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := routes.SetupRouter(handlers, cfg)

	s := &Server{
		Config: &cfg,
		DB:     db,
		Router: router,
		httpServer: &http.Server{
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	return s
}

// Start runs the server
func (s *Server) Start(port string) error {
	s.httpServer.Addr = port
	slog.Info("🚀 Server running", "port", port, "mode", s.Config.RunMode)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown() error {
	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
		return err
	}
	slog.Info("Server shut down successfully")
	return nil
}
