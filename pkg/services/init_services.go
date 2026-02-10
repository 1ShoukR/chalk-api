package services

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/events"
	"chalk-api/pkg/repositories"
)

// InitializeServices initializes all services
func InitializeServices(repos *repositories.RepositoriesCollection, cfg config.Environment) (*ServicesCollection, error) {
	return &ServicesCollection{
		Events: events.NewPublisher(repos.Outbox),
	}, nil
}

// ServicesCollection contains all the services
type ServicesCollection struct {
	Events *events.Publisher
}
