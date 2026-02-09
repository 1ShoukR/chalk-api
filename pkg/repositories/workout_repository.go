package repositories

import (
	"chalk-api/pkg/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type WorkoutRepository struct {
	db *gorm.DB
}

func NewWorkoutRepository(db *gorm.DB) *WorkoutRepository {
	return &WorkoutRepository{db: db}
}

// Create creates a workout with all exercises in one transaction (deep copy from template)
func (r *WorkoutRepository) Create(ctx context.Context, workout *models.Workout) error {
	return r.db.WithContext(ctx).Create(workout).Error
}

func (r *WorkoutRepository) GetByID(ctx context.Context, id uint) (*models.Workout, error) {
	var workout models.Workout
	err := r.db.WithContext(ctx).
		Preload("Exercises", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_index ASC")
		}).
		Preload("Exercises.Exercise").
		Preload("Exercises.Logs", func(db *gorm.DB) *gorm.DB {
			return db.Order("set_number ASC")
		}).
		First(&workout, id).Error
	if err != nil {
		return nil, err
	}
	return &workout, nil
}

func (r *WorkoutRepository) GetByClientAndDate(ctx context.Context, clientID uint, date string) (*models.Workout, error) {
	var workout models.Workout
	err := r.db.WithContext(ctx).
		Preload("Exercises", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_index ASC")
		}).
		Preload("Exercises.Exercise").
		Preload("Exercises.Logs", func(db *gorm.DB) *gorm.DB {
			return db.Order("set_number ASC")
		}).
		Where("client_id = ? AND scheduled_date = ?", clientID, date).
		First(&workout).Error
	if err != nil {
		return nil, err
	}
	return &workout, nil
}

func (r *WorkoutRepository) ListByClient(ctx context.Context, clientID uint, limit, offset int) ([]models.Workout, int64, error) {
	var workouts []models.Workout
	var total int64

	query := r.db.WithContext(ctx).Where("client_id = ?", clientID)

	if err := query.Model(&models.Workout{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("scheduled_date DESC").
		Limit(limit).Offset(offset).
		Find(&workouts).Error

	return workouts, total, err
}

func (r *WorkoutRepository) Update(ctx context.Context, workout *models.Workout) error {
	return r.db.WithContext(ctx).Save(workout).Error
}

func (r *WorkoutRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Workout{}, id).Error
}

// --- Status Updates ---

func (r *WorkoutRepository) StartWorkout(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Workout{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     "in_progress",
			"started_at": now,
		}).Error
}

func (r *WorkoutRepository) CompleteWorkout(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Workout{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": now,
		}).Error
}

func (r *WorkoutRepository) SkipWorkout(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Workout{}).
		Where("id = ?", id).
		Update("status", "skipped").Error
}

// --- Exercise Completion ---

func (r *WorkoutRepository) MarkExerciseCompleted(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.WorkoutExercise{}).
		Where("id = ?", id).
		Update("is_completed", true).Error
}

func (r *WorkoutRepository) SkipExercise(ctx context.Context, id uint, reason string) error {
	return r.db.WithContext(ctx).
		Model(&models.WorkoutExercise{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_completed":   false,
			"skipped_reason": reason,
		}).Error
}

// --- Workout Logs ---

func (r *WorkoutRepository) CreateLog(ctx context.Context, log *models.WorkoutLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *WorkoutRepository) UpdateLog(ctx context.Context, log *models.WorkoutLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

func (r *WorkoutRepository) ListLogsByExercise(ctx context.Context, workoutExerciseID uint) ([]models.WorkoutLog, error) {
	var logs []models.WorkoutLog
	err := r.db.WithContext(ctx).
		Where("workout_exercise_id = ?", workoutExerciseID).
		Order("set_number ASC").
		Find(&logs).Error
	return logs, err
}
