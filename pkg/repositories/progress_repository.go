package repositories

import (
	"chalk-api/pkg/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type ProgressRepository struct {
	db *gorm.DB
}

func NewProgressRepository(db *gorm.DB) *ProgressRepository {
	return &ProgressRepository{db: db}
}

// --- Metrics ---

func (r *ProgressRepository) CreateMetric(ctx context.Context, metric *models.BodyMetric) error {
	return r.db.WithContext(ctx).Create(metric).Error
}

func (r *ProgressRepository) ListMetrics(ctx context.Context, clientID uint, metricType string, startDate, endDate time.Time) ([]models.BodyMetric, error) {
	var metrics []models.BodyMetric

	query := r.db.WithContext(ctx).
		Where("client_id = ?", clientID)

	if metricType != "" {
		query = query.Where("metric_type = ?", metricType)
	}
	if !startDate.IsZero() {
		query = query.Where("recorded_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("recorded_at <= ?", endDate)
	}

	err := query.Order("recorded_at DESC").Find(&metrics).Error
	return metrics, err
}

// GetLatestMetric returns the most recent value for a given metric type
func (r *ProgressRepository) GetLatestMetric(ctx context.Context, clientID uint, metricType string) (*models.BodyMetric, error) {
	var metric models.BodyMetric
	err := r.db.WithContext(ctx).
		Where("client_id = ? AND metric_type = ?", clientID, metricType).
		Order("recorded_at DESC").
		First(&metric).Error
	if err != nil {
		return nil, err
	}
	return &metric, nil
}

func (r *ProgressRepository) DeleteMetric(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.BodyMetric{}, id).Error
}

// --- Photos ---

func (r *ProgressRepository) CreatePhoto(ctx context.Context, photo *models.ProgressPhoto) error {
	return r.db.WithContext(ctx).Create(photo).Error
}

func (r *ProgressRepository) ListPhotos(ctx context.Context, clientID uint, photoType string, startDate, endDate time.Time) ([]models.ProgressPhoto, error) {
	var photos []models.ProgressPhoto

	query := r.db.WithContext(ctx).
		Where("client_id = ?", clientID)

	if photoType != "" {
		query = query.Where("photo_type = ?", photoType)
	}
	if !startDate.IsZero() {
		query = query.Where("taken_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("taken_at <= ?", endDate)
	}

	err := query.Order("taken_at DESC").Find(&photos).Error
	return photos, err
}

func (r *ProgressRepository) DeletePhoto(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ProgressPhoto{}, id).Error
}
