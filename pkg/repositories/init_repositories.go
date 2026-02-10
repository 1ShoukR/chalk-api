package repositories

import "gorm.io/gorm"

type RepositoriesCollection struct {
	User         *UserRepository
	Auth         *AuthRepository
	Coach        *CoachRepository
	Client       *ClientRepository
	Subscription *SubscriptionRepository
	Exercise     *ExerciseRepository
	Template     *TemplateRepository
	Workout      *WorkoutRepository
	Session      *SessionRepository
	Nutrition    *NutritionRepository
	Progress     *ProgressRepository
	Message      *MessageRepository
	Outbox       *OutboxRepository
}

func InitializeRepositories(db *gorm.DB) (*RepositoriesCollection, error) {
	return &RepositoriesCollection{
		User:         NewUserRepository(db),
		Auth:         NewAuthRepository(db),
		Coach:        NewCoachRepository(db),
		Client:       NewClientRepository(db),
		Subscription: NewSubscriptionRepository(db),
		Exercise:     NewExerciseRepository(db),
		Template:     NewTemplateRepository(db),
		Workout:      NewWorkoutRepository(db),
		Session:      NewSessionRepository(db),
		Nutrition:    NewNutritionRepository(db),
		Progress:     NewProgressRepository(db),
		Message:      NewMessageRepository(db),
		Outbox:       NewOutboxRepository(db),
	}, nil
}
