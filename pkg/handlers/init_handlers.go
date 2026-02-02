package handlers

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/repositories"
	"chalk-api/pkg/services"
)

// InitializeHandlers initializes all the handlers
func InitializeHandlers(services *services.ServicesCollection, repos *repositories.RepositoriesCollection, cfg config.Environment) (*HandlersCollection, error) {
	return &HandlersCollection{
		// Add handlers here as you create them
		// Example:
		// UserHandler: NewUserHandler(services.UserService),
	}, nil
}

// HandlersCollection contains all the handlers
type HandlersCollection struct {
	// Add handler fields here as you create them
	// Example:
	// UserHandler *UserHandler
}
