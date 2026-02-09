package repositories

import (
	"chalk-api/pkg/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// --- Availability ---

// SetAvailability replaces all recurring slots for a coach in a transaction
func (r *SessionRepository) SetAvailability(ctx context.Context, coachID uint, slots []models.CoachAvailability) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("coach_id = ?", coachID).Delete(&models.CoachAvailability{}).Error; err != nil {
			return err
		}
		if len(slots) == 0 {
			return nil
		}
		return tx.Create(&slots).Error
	})
}

func (r *SessionRepository) GetAvailability(ctx context.Context, coachID uint) ([]models.CoachAvailability, error) {
	var slots []models.CoachAvailability
	err := r.db.WithContext(ctx).
		Where("coach_id = ? AND is_active = ?", coachID, true).
		Order("day_of_week ASC, start_time ASC").
		Find(&slots).Error
	return slots, err
}

func (r *SessionRepository) UpdateAvailabilitySlot(ctx context.Context, slot *models.CoachAvailability) error {
	return r.db.WithContext(ctx).Save(slot).Error
}

func (r *SessionRepository) DeleteAvailabilitySlot(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.CoachAvailability{}, id).Error
}

// --- Overrides ---

func (r *SessionRepository) CreateOverride(ctx context.Context, override *models.CoachAvailabilityOverride) error {
	return r.db.WithContext(ctx).Create(override).Error
}

func (r *SessionRepository) ListOverrides(ctx context.Context, coachID uint, startDate, endDate string) ([]models.CoachAvailabilityOverride, error) {
	var overrides []models.CoachAvailabilityOverride
	err := r.db.WithContext(ctx).
		Where("coach_id = ? AND date >= ? AND date <= ?", coachID, startDate, endDate).
		Order("date ASC").
		Find(&overrides).Error
	return overrides, err
}

func (r *SessionRepository) DeleteOverride(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.CoachAvailabilityOverride{}, id).Error
}

// --- Session Types ---

func (r *SessionRepository) CreateSessionType(ctx context.Context, st *models.SessionType) error {
	return r.db.WithContext(ctx).Create(st).Error
}

func (r *SessionRepository) ListSessionTypes(ctx context.Context, coachID uint) ([]models.SessionType, error) {
	var types []models.SessionType
	err := r.db.WithContext(ctx).
		Where("coach_id = ? AND is_active = ?", coachID, true).
		Order("name ASC").
		Find(&types).Error
	return types, err
}

func (r *SessionRepository) UpdateSessionType(ctx context.Context, st *models.SessionType) error {
	return r.db.WithContext(ctx).Save(st).Error
}

func (r *SessionRepository) DeleteSessionType(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.SessionType{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// --- Sessions ---

func (r *SessionRepository) CreateSession(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *SessionRepository) GetSession(ctx context.Context, id uint) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).
		Preload("Client.User.Profile").
		Preload("SessionType").
		First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// ListSessions returns sessions for a coach or client within a date range
func (r *SessionRepository) ListSessions(ctx context.Context, coachID, clientID uint, startDate, endDate time.Time) ([]models.Session, error) {
	var sessions []models.Session

	query := r.db.WithContext(ctx).
		Preload("Client.User.Profile").
		Preload("SessionType").
		Where("scheduled_at >= ? AND scheduled_at <= ?", startDate, endDate)

	if coachID > 0 {
		query = query.Where("coach_id = ?", coachID)
	}
	if clientID > 0 {
		query = query.Where("client_id = ?", clientID)
	}

	err := query.Order("scheduled_at ASC").Find(&sessions).Error
	return sessions, err
}

func (r *SessionRepository) UpdateSession(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *SessionRepository) CompleteSession(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": now,
		}).Error
}

func (r *SessionRepository) CancelSession(ctx context.Context, id uint, cancelledBy, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":              "cancelled",
			"cancelled_at":        now,
			"cancelled_by":        cancelledBy,
			"cancellation_reason": reason,
		}).Error
}

func (r *SessionRepository) MarkNoShow(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("id = ?", id).
		Update("status", "no_show").Error
}
