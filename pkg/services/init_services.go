package services

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/repositories"
)

// InitializeServices initializes all services
func InitializeServices(repos *repositories.RepositoriesCollection, cfg config.Environment) (*ServicesCollection, error) {
	return &ServicesCollection{
		// Add services here as you create them
		// Example:
		// UserService: NewUserService(repos.UserRepository),
	}, nil
}

// ServicesCollection contains all the services
type ServicesCollection struct {
	// Add service fields here as you create them
	// Example:
	// UserService *UserService
}
