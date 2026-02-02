package clients

import (
	"chalk-api/pkg/config"
)

// ClientsCollection contains all external API clients
type ClientsCollection struct {
	// Add client fields here as you create them
	// Example:
	// RedisClient  *redis.Client
}

// InitializeClients initializes all external API clients
func InitializeClients(cfg config.Environment) (*ClientsCollection, error) {
	return &ClientsCollection{
		// Initialize clients here
	}, nil
}

// CloseAll closes all client connections
func (c *ClientsCollection) CloseAll() {
	// Close all clients here
}
