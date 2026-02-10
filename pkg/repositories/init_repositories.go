package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type RepositoriesCollection struct {
	db *gorm.DB

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
	return newRepositoriesCollection(db), nil
}

func newRepositoriesCollection(db *gorm.DB) *RepositoriesCollection {
	return &RepositoriesCollection{
		db: db,

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
	}
}

// WithTransaction runs fn inside a single DB transaction and provides tx-scoped repositories.
func (r *RepositoriesCollection) WithTransaction(
	ctx context.Context,
	fn func(tx *gorm.DB, txRepos *RepositoriesCollection) error,
) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("repositories collection is not initialized")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepos := newRepositoriesCollection(tx)
		return fn(tx, txRepos)
	})
}
