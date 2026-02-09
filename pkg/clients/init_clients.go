package clients

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/external"
)

// ClientsCollection wraps external integrations
// Deprecated: Use pkg/external.Collection directly
type ClientsCollection struct {
	External *external.Collection
}

// InitializeClients initializes all external API integrations
func InitializeClients(cfg config.Environment) (*ClientsCollection, error) {
	return &ClientsCollection{
		External: external.Initialize(cfg),
	}, nil
}

// CloseAll closes all client connections
func (c *ClientsCollection) CloseAll() {
	// External APIs don't need cleanup (stateless HTTP)
}
