package repositories

import (
	"chalk-api/pkg/models"
	"context"

	"gorm.io/gorm"
)

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// Create creates a template with its exercises in a single transaction
func (r *TemplateRepository) Create(ctx context.Context, template *models.WorkoutTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

func (r *TemplateRepository) GetByID(ctx context.Context, id uint) (*models.WorkoutTemplate, error) {
	var template models.WorkoutTemplate
	err := r.db.WithContext(ctx).
		Preload("Exercises", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_index ASC")
		}).
		Preload("Exercises.Exercise").
		First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) ListByCoach(ctx context.Context, coachID uint, limit, offset int) ([]models.WorkoutTemplate, int64, error) {
	var templates []models.WorkoutTemplate
	var total int64

	query := r.db.WithContext(ctx).
		Where("coach_id = ? AND is_active = ?", coachID, true)

	if err := query.Model(&models.WorkoutTemplate{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("updated_at DESC").
		Limit(limit).Offset(offset).
		Find(&templates).Error

	return templates, total, err
}

func (r *TemplateRepository) Update(ctx context.Context, template *models.WorkoutTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

func (r *TemplateRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.WorkoutTemplate{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// --- Template Exercises ---

func (r *TemplateRepository) AddExercise(ctx context.Context, exercise *models.WorkoutTemplateExercise) error {
	return r.db.WithContext(ctx).Create(exercise).Error
}

func (r *TemplateRepository) UpdateExercise(ctx context.Context, exercise *models.WorkoutTemplateExercise) error {
	return r.db.WithContext(ctx).Save(exercise).Error
}

func (r *TemplateRepository) RemoveExercise(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.WorkoutTemplateExercise{}, id).Error
}

// ReorderExercises updates order_index for multiple exercises in a single transaction
func (r *TemplateRepository) ReorderExercises(ctx context.Context, templateID uint, orderMap map[uint]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for exerciseID, newOrder := range orderMap {
			if err := tx.Model(&models.WorkoutTemplateExercise{}).
				Where("id = ? AND template_id = ?", exerciseID, templateID).
				Update("order_index", newOrder).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ReplaceExercises replaces all template exercises in a single transaction.
func (r *TemplateRepository) ReplaceExercises(ctx context.Context, templateID uint, exercises []models.WorkoutTemplateExercise) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("template_id = ?", templateID).Delete(&models.WorkoutTemplateExercise{}).Error; err != nil {
			return err
		}

		if len(exercises) == 0 {
			return nil
		}

		for i := range exercises {
			exercises[i].TemplateID = templateID
		}

		return tx.Create(&exercises).Error
	})
}
