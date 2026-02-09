package repositories

import (
	"chalk-api/pkg/models"
	"context"

	"gorm.io/gorm"
)

type CoachRepository struct {
	db *gorm.DB
}

func NewCoachRepository(db *gorm.DB) *CoachRepository {
	return &CoachRepository{db: db}
}

func (r *CoachRepository) Create(ctx context.Context, profile *models.CoachProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *CoachRepository) GetByID(ctx context.Context, id uint) (*models.CoachProfile, error) {
	var profile models.CoachProfile
	err := r.db.WithContext(ctx).
		Preload("Certifications").
		Preload("Locations").
		Preload("Stats").
		First(&profile, id).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *CoachRepository) GetByUserID(ctx context.Context, userID uint) (*models.CoachProfile, error) {
	var profile models.CoachProfile
	err := r.db.WithContext(ctx).
		Preload("Certifications").
		Preload("Locations").
		Preload("Stats").
		Where("user_id = ?", userID).
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *CoachRepository) Update(ctx context.Context, profile *models.CoachProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

// --- Certifications ---

func (r *CoachRepository) AddCertification(ctx context.Context, cert *models.Certification) error {
	return r.db.WithContext(ctx).Create(cert).Error
}

func (r *CoachRepository) ListCertifications(ctx context.Context, coachID uint) ([]models.Certification, error) {
	var certs []models.Certification
	err := r.db.WithContext(ctx).
		Where("coach_id = ?", coachID).
		Order("created_at DESC").
		Find(&certs).Error
	return certs, err
}

func (r *CoachRepository) UpdateCertification(ctx context.Context, cert *models.Certification) error {
	return r.db.WithContext(ctx).Save(cert).Error
}

func (r *CoachRepository) RemoveCertification(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Certification{}, id).Error
}

// --- Locations ---

func (r *CoachRepository) AddLocation(ctx context.Context, location *models.CoachLocation) error {
	return r.db.WithContext(ctx).Create(location).Error
}

func (r *CoachRepository) ListLocations(ctx context.Context, coachID uint) ([]models.CoachLocation, error) {
	var locations []models.CoachLocation
	err := r.db.WithContext(ctx).
		Where("coach_id = ? AND is_active = ?", coachID, true).
		Order("is_primary DESC").
		Find(&locations).Error
	return locations, err
}

func (r *CoachRepository) UpdateLocation(ctx context.Context, location *models.CoachLocation) error {
	return r.db.WithContext(ctx).Save(location).Error
}

func (r *CoachRepository) RemoveLocation(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.CoachLocation{}, id).Error
}

// SetPrimaryLocation clears existing primary and sets a new one in a transaction
func (r *CoachRepository) SetPrimaryLocation(ctx context.Context, coachID uint, locationID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.CoachLocation{}).
			Where("coach_id = ?", coachID).
			Update("is_primary", false).Error; err != nil {
			return err
		}
		return tx.Model(&models.CoachLocation{}).
			Where("id = ? AND coach_id = ?", locationID, coachID).
			Update("is_primary", true).Error
	})
}

// --- Stats ---

func (r *CoachRepository) GetStats(ctx context.Context, coachID uint) (*models.CoachStats, error) {
	var stats models.CoachStats
	err := r.db.WithContext(ctx).
		Where("coach_id = ?", coachID).
		First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *CoachRepository) UpdateStats(ctx context.Context, stats *models.CoachStats) error {
	return r.db.WithContext(ctx).Save(stats).Error
}

func (r *CoachRepository) IncrementStat(ctx context.Context, coachID uint, field string, amount int) error {
	return r.db.WithContext(ctx).
		Model(&models.CoachStats{}).
		Where("coach_id = ?", coachID).
		Update(field, gorm.Expr(field+" + ?", amount)).Error
}
