package handlers

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/repositories"
	"chalk-api/pkg/services"
)

// InitializeHandlers initializes all the handlers
func InitializeHandlers(services *services.ServicesCollection, repos *repositories.RepositoriesCollection, cfg config.Environment) (*HandlersCollection, error) {
	return &HandlersCollection{
		Auth:    NewAuthHandler(services.Auth),
		User:    NewUserHandler(services.User),
		Coach:   NewCoachHandler(services.Coach),
		Invite:  NewInviteHandler(services.Coach),
		Workout: NewWorkoutHandler(services.Workout),
		Message: NewMessageHandler(services.Message),
	}, nil
}

// HandlersCollection contains all the handlers
type HandlersCollection struct {
	Auth    *AuthHandler
	User    *UserHandler
	Coach   *CoachHandler
	Invite  *InviteHandler
	Workout *WorkoutHandler
	Message *MessageHandler
}
