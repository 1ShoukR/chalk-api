package services

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/events"
	"chalk-api/pkg/external"
	"chalk-api/pkg/repositories"
)

// InitializeServices initializes all services
func InitializeServices(
	repos *repositories.RepositoriesCollection,
	integrations *external.Collection,
	cfg config.Environment,
) (*ServicesCollection, error) {
	eventsPublisher := events.NewPublisher(repos.Outbox)

	if integrations == nil {
		integrations = &external.Collection{}
	}

	return &ServicesCollection{
		Events:       eventsPublisher,
		Auth:         NewAuthService(repos.User, repos.Auth, cfg.JWTSecret, cfg.JWTExpirationHours),
		User:         NewUserService(repos.User),
		Coach:        NewCoachService(repos, eventsPublisher),
		Session:      NewSessionService(repos, eventsPublisher),
		Workout:      NewWorkoutService(repos, eventsPublisher),
		Message:      NewMessageService(repos, eventsPublisher),
		Subscription: NewSubscriptionService(repos, integrations.RevenueCat),
	}, nil
}

// ServicesCollection contains all the services
type ServicesCollection struct {
	Events       *events.Publisher
	Auth         *AuthService
	User         *UserService
	Coach        *CoachService
	Session      *SessionService
	Workout      *WorkoutService
	Message      *MessageService
	Subscription *SubscriptionService
}
