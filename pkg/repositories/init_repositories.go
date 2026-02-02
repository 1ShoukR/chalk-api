package repositories

import (
	"gorm.io/gorm"
)

// InitializeRepositories initializes all repositories
func InitializeRepositories(db *gorm.DB) (*RepositoriesCollection, error) {
	return &RepositoriesCollection{
		// Add repositories here as you create them
		// Example:
		// UserRepository: NewUserRepository(db),
	}, nil
}

// RepositoriesCollection contains all the repositories
type RepositoriesCollection struct {
	// Add repository fields here as you create them
	// Example:
	// UserRepository *UserRepository
}
