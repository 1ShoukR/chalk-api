package services

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/events"
	"chalk-api/pkg/repositories"
)

// InitializeServices initializes all services
func InitializeServices(repos *repositories.RepositoriesCollection, cfg config.Environment) (*ServicesCollection, error) {
	eventsPublisher := events.NewPublisher(repos.Outbox)

	return &ServicesCollection{
		Events:  eventsPublisher,
		Auth:    NewAuthService(repos.User, repos.Auth, cfg.JWTSecret, cfg.JWTExpirationHours),
		User:    NewUserService(repos.User),
		Coach:   NewCoachService(repos.Coach, repos.Client, eventsPublisher),
		Workout: NewWorkoutService(repos.Template, repos.Workout, repos.Coach, repos.Client, eventsPublisher),
	}, nil
}

// ServicesCollection contains all the services
type ServicesCollection struct {
	Events  *events.Publisher
	Auth    *AuthService
	User    *UserService
	Coach   *CoachService
	Workout *WorkoutService
}
