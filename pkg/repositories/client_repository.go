package repositories

import (
	"chalk-api/pkg/models"
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ClientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

func (r *ClientRepository) Create(ctx context.Context, profile *models.ClientProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *ClientRepository) GetByID(ctx context.Context, id uint) (*models.ClientProfile, error) {
	var profile models.ClientProfile
	err := r.db.WithContext(ctx).
		Preload("User.Profile").
		Preload("IntakeForm").
		First(&profile, id).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *ClientRepository) GetByUserAndCoach(ctx context.Context, userID, coachID uint) (*models.ClientProfile, error) {
	var profile models.ClientProfile
	err := r.db.WithContext(ctx).
		Preload("User.Profile").
		Where("user_id = ? AND coach_id = ?", userID, coachID).
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// ListByCoach returns paginated clients for a coach, filterable by status
func (r *ClientRepository) ListByCoach(ctx context.Context, coachID uint, status string, limit, offset int) ([]models.ClientProfile, int64, error) {
	var clients []models.ClientProfile
	var total int64

	query := r.db.WithContext(ctx).
		Where("coach_id = ?", coachID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Model(&models.ClientProfile{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("User.Profile").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&clients).Error

	return clients, total, err
}

// ListByUser returns all coach relationships for a user
func (r *ClientRepository) ListByUser(ctx context.Context, userID uint) ([]models.ClientProfile, error) {
	var clients []models.ClientProfile
	err := r.db.WithContext(ctx).
		Preload("Coach.User.Profile").
		Where("user_id = ?", userID).
		Find(&clients).Error
	return clients, err
}

func (r *ClientRepository) Update(ctx context.Context, profile *models.ClientProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

func (r *ClientRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.ClientProfile{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *ClientRepository) UpdatePrivateNotes(ctx context.Context, id uint, notes string) error {
	return r.db.WithContext(ctx).
		Model(&models.ClientProfile{}).
		Where("id = ?", id).
		Update("private_notes", notes).Error
}

// --- Invite Codes ---

func (r *ClientRepository) CreateInviteCode(ctx context.Context, code *models.InviteCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

func (r *ClientRepository) GetInviteCode(ctx context.Context, code string) (*models.InviteCode, error) {
	var invite models.InviteCode
	err := r.db.WithContext(ctx).
		Where("code = ? AND is_active = ? AND expires_at > ? AND used_by IS NULL", code, true, time.Now()).
		First(&invite).Error
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *ClientRepository) GetInviteCodeByID(ctx context.Context, id uint) (*models.InviteCode, error) {
	var invite models.InviteCode
	err := r.db.WithContext(ctx).First(&invite, id).Error
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *ClientRepository) MarkInviteUsed(ctx context.Context, id uint, userID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.InviteCode{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"used_by": userID,
			"used_at": now,
		}).Error
}

// AcceptInvite creates the coach-client relationship and marks the invite as used in one transaction.
// Returns alreadyConnected=true when the relationship already exists.
func (r *ClientRepository) AcceptInvite(ctx context.Context, invite *models.InviteCode, userID uint) (*models.ClientProfile, bool, error) {
	var result models.ClientProfile
	alreadyConnected := false
	now := time.Now()

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.ClientProfile
		err := tx.Where("user_id = ? AND coach_id = ?", userID, invite.CoachID).First(&existing).Error
		if err == nil {
			alreadyConnected = true
			result = existing
			return tx.Model(&models.InviteCode{}).
				Where("id = ? AND used_by IS NULL", invite.ID).
				Updates(map[string]any{
					"used_by": userID,
					"used_at": now,
				}).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		invitedAt := invite.CreatedAt
		profile := models.ClientProfile{
			UserID:    userID,
			CoachID:   invite.CoachID,
			Status:    "active",
			InvitedAt: &invitedAt,
			JoinedAt:  &now,
		}
		if err := tx.Create(&profile).Error; err != nil {
			// Handle race where another request creates the relation first.
			if isDuplicateKeyError(err) {
				if getErr := tx.Where("user_id = ? AND coach_id = ?", userID, invite.CoachID).First(&existing).Error; getErr == nil {
					alreadyConnected = true
					result = existing
				} else {
					return getErr
				}
			} else {
				return err
			}
		} else {
			result = profile
		}

		return tx.Model(&models.InviteCode{}).
			Where("id = ? AND used_by IS NULL", invite.ID).
			Updates(map[string]any{
				"used_by": userID,
				"used_at": now,
			}).Error
	})
	if err != nil {
		return nil, false, err
	}

	if err := r.db.WithContext(ctx).
		Preload("User.Profile").
		Preload("Coach.User.Profile").
		First(&result, result.ID).Error; err != nil {
		return nil, alreadyConnected, err
	}

	return &result, alreadyConnected, nil
}

func (r *ClientRepository) ListInviteCodes(ctx context.Context, coachID uint) ([]models.InviteCode, error) {
	var codes []models.InviteCode
	err := r.db.WithContext(ctx).
		Where("coach_id = ?", coachID).
		Order("created_at DESC").
		Find(&codes).Error
	return codes, err
}

func (r *ClientRepository) DeactivateInviteCode(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.InviteCode{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}

// --- Intake Form ---

func (r *ClientRepository) CreateIntakeForm(ctx context.Context, form *models.ClientIntakeForm) error {
	return r.db.WithContext(ctx).Create(form).Error
}

func (r *ClientRepository) GetIntakeForm(ctx context.Context, clientID uint) (*models.ClientIntakeForm, error) {
	var form models.ClientIntakeForm
	err := r.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		First(&form).Error
	if err != nil {
		return nil, err
	}
	return &form, nil
}

func (r *ClientRepository) UpdateIntakeForm(ctx context.Context, form *models.ClientIntakeForm) error {
	return r.db.WithContext(ctx).Save(form).Error
}
