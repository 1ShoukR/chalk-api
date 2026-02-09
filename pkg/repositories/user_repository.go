package repositories

import (
	"chalk-api/pkg/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a user and profile in a single transaction
func (r *UserRepository) Create(ctx context.Context, user *models.User, profile *models.Profile) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		profile.UserID = user.ID
		return tx.Create(profile).Error
	})
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("Profile").
		First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("Profile").
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) UpdateProfile(ctx context.Context, profile *models.Profile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_login_at", now).Error
}

// --- OAuth Providers ---

func (r *UserRepository) AddOAuthProvider(ctx context.Context, provider *models.OAuthProvider) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

func (r *UserRepository) GetOAuthProvider(ctx context.Context, provider string, providerUserID string) (*models.OAuthProvider, error) {
	var oauth models.OAuthProvider
	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_user_id = ?", provider, providerUserID).
		First(&oauth).Error
	if err != nil {
		return nil, err
	}
	return &oauth, nil
}

func (r *UserRepository) ListOAuthProviders(ctx context.Context, userID uint) ([]models.OAuthProvider, error) {
	var providers []models.OAuthProvider
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&providers).Error
	return providers, err
}

func (r *UserRepository) RemoveOAuthProvider(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.OAuthProvider{}, id).Error
}

// --- Device Tokens ---

func (r *UserRepository) AddDeviceToken(ctx context.Context, token *models.DeviceToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *UserRepository) GetDeviceTokens(ctx context.Context, userID uint) ([]models.DeviceToken, error) {
	var tokens []models.DeviceToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Find(&tokens).Error
	return tokens, err
}

func (r *UserRepository) DeactivateDeviceToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Model(&models.DeviceToken{}).
		Where("token = ?", token).
		Update("is_active", false).Error
}
