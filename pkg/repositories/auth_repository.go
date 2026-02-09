package repositories

import (
	"chalk-api/pkg/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// --- Refresh Tokens ---

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *AuthRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND revoked = ? AND expires_at > ?", tokenHash, false, time.Now()).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"revoked":    true,
			"revoked_at": now,
		}).Error
}

// RevokeAllUserTokens revokes every refresh token for a user (logout everywhere)
func (r *AuthRepository) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Updates(map[string]interface{}{
			"revoked":    true,
			"revoked_at": now,
		}).Error
}

func (r *AuthRepository) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ? OR revoked = ?", time.Now(), true).
		Delete(&models.RefreshToken{})
	return result.RowsAffected, result.Error
}

// --- Password Resets ---

func (r *AuthRepository) CreatePasswordReset(ctx context.Context, reset *models.PasswordReset) error {
	return r.db.WithContext(ctx).Create(reset).Error
}

func (r *AuthRepository) GetPasswordReset(ctx context.Context, tokenHash string) (*models.PasswordReset, error) {
	var reset models.PasswordReset
	err := r.db.WithContext(ctx).
		Where("token = ? AND used = ? AND expires_at > ?", tokenHash, false, time.Now()).
		First(&reset).Error
	if err != nil {
		return nil, err
	}
	return &reset, nil
}

func (r *AuthRepository) MarkPasswordResetUsed(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.PasswordReset{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"used":    true,
			"used_at": now,
		}).Error
}

func (r *AuthRepository) CleanupExpiredResets(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.PasswordReset{})
	return result.RowsAffected, result.Error
}

// --- Email Verification ---

func (r *AuthRepository) CreateEmailVerification(ctx context.Context, verification *models.EmailVerification) error {
	return r.db.WithContext(ctx).Create(verification).Error
}

func (r *AuthRepository) GetEmailVerification(ctx context.Context, tokenHash string) (*models.EmailVerification, error) {
	var verification models.EmailVerification
	err := r.db.WithContext(ctx).
		Where("token = ? AND used = ? AND expires_at > ?", tokenHash, false, time.Now()).
		First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

func (r *AuthRepository) MarkEmailVerified(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.EmailVerification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"used":    true,
			"used_at": now,
		}).Error
}

// --- Magic Links ---

func (r *AuthRepository) CreateMagicLink(ctx context.Context, link *models.MagicLink) error {
	return r.db.WithContext(ctx).Create(link).Error
}

func (r *AuthRepository) GetMagicLink(ctx context.Context, tokenHash string) (*models.MagicLink, error) {
	var link models.MagicLink
	err := r.db.WithContext(ctx).
		Where("token = ? AND used = ? AND expires_at > ?", tokenHash, false, time.Now()).
		First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *AuthRepository) MarkMagicLinkUsed(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.MagicLink{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"used":    true,
			"used_at": now,
		}).Error
}

func (r *AuthRepository) CleanupExpiredMagicLinks(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.MagicLink{})
	return result.RowsAffected, result.Error
}
