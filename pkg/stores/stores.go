package stores

import (
	"chalk-api/pkg/config"
)

// StoresCollection contains all runtime stores (e.g., Redis-backed caches)
type StoresCollection struct {
	// Add store fields here as you create them
	// Example:
	// SessionStore *SessionStore
	// CacheStore   *CacheStore
}

// InitializeStores initializes all runtime stores
func InitializeStores(cfg config.Environment) (*StoresCollection, error) {
	return &StoresCollection{
		// Initialize stores here
	}, nil
}
