package repositories

import (
	"chalk-api/pkg/models"
	"context"

	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID uint) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

// IsActive checks if a user has an active subscription without loading the full record
func (r *SubscriptionRepository) IsActive(ctx context.Context, userID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("user_id = ? AND status = ?", userID, "active").
		Count(&count).Error
	return count > 0, err
}

// --- Events ---

func (r *SubscriptionRepository) CreateEvent(ctx context.Context, event *models.SubscriptionEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// GetEventByRevenueCatID prevents duplicate webhook processing
func (r *SubscriptionRepository) GetEventByRevenueCatID(ctx context.Context, eventID string) (*models.SubscriptionEvent, error) {
	var event models.SubscriptionEvent
	err := r.db.WithContext(ctx).
		Where("revenuecat_event_id = ?", eventID).
		First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *SubscriptionRepository) ListEvents(ctx context.Context, subscriptionID uint) ([]models.SubscriptionEvent, error) {
	var events []models.SubscriptionEvent
	err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("processed_at DESC").
		Find(&events).Error
	return events, err
}
