package repositories

import (
	"chalk-api/pkg/models"
	"context"

	"gorm.io/gorm"
)

type ExerciseRepository struct {
	db *gorm.DB
}

func NewExerciseRepository(db *gorm.DB) *ExerciseRepository {
	return &ExerciseRepository{db: db}
}

func (r *ExerciseRepository) Create(ctx context.Context, exercise *models.Exercise) error {
	return r.db.WithContext(ctx).Create(exercise).Error
}

func (r *ExerciseRepository) GetByID(ctx context.Context, id uint) (*models.Exercise, error) {
	var exercise models.Exercise
	err := r.db.WithContext(ctx).First(&exercise, id).Error
	if err != nil {
		return nil, err
	}
	return &exercise, nil
}

func (r *ExerciseRepository) Update(ctx context.Context, exercise *models.Exercise) error {
	return r.db.WithContext(ctx).Save(exercise).Error
}

func (r *ExerciseRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Exercise{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// List returns paginated exercises with optional filters
func (r *ExerciseRepository) List(ctx context.Context, category, difficulty string, limit, offset int) ([]models.Exercise, int64, error) {
	var exercises []models.Exercise
	var total int64

	query := r.db.WithContext(ctx).Where("is_active = ?", true)

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}

	if err := query.Model(&models.Exercise{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("name ASC").
		Limit(limit).Offset(offset).
		Find(&exercises).Error

	return exercises, total, err
}

// Search performs text search on exercise name
func (r *ExerciseRepository) Search(ctx context.Context, query string, limit, offset int) ([]models.Exercise, int64, error) {
	var exercises []models.Exercise
	var total int64

	dbQuery := r.db.WithContext(ctx).
		Where("is_active = ? AND name ILIKE ?", true, "%"+query+"%")

	if err := dbQuery.Model(&models.Exercise{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := dbQuery.
		Order("name ASC").
		Limit(limit).Offset(offset).
		Find(&exercises).Error

	return exercises, total, err
}

// ListByCoach returns a coach's custom exercises
func (r *ExerciseRepository) ListByCoach(ctx context.Context, coachID uint) ([]models.Exercise, error) {
	var exercises []models.Exercise
	err := r.db.WithContext(ctx).
		Where("coach_id = ? AND is_active = ?", coachID, true).
		Order("name ASC").
		Find(&exercises).Error
	return exercises, err
}

// ListSystem returns all system exercises available to every coach
func (r *ExerciseRepository) ListSystem(ctx context.Context) ([]models.Exercise, error) {
	var exercises []models.Exercise
	err := r.db.WithContext(ctx).
		Where("is_system = ? AND is_active = ?", true, true).
		Order("name ASC").
		Find(&exercises).Error
	return exercises, err
}

// GetByExternalID checks if we already cached an exercise from a third-party API
func (r *ExerciseRepository) GetByExternalID(ctx context.Context, source, externalID string) (*models.Exercise, error) {
	var exercise models.Exercise
	err := r.db.WithContext(ctx).
		Where("source = ? AND external_id = ?", source, externalID).
		First(&exercise).Error
	if err != nil {
		return nil, err
	}
	return &exercise, nil
}
