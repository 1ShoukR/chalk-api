package stores

import (
	"chalk-api/pkg/models"
	"time"
)

// CoachStore handles coach profile caching
type CoachStore struct {
	redis *RedisClient
}

// Cache TTLs for coach data
const (
	CoachProfileTTL     = 15 * time.Minute
	CoachStatsTTL       = 30 * time.Minute
	CoachAvailabilityTTL = 5 * time.Minute
)

// NewCoachStore creates a new coach store
func NewCoachStore(redis *RedisClient) *CoachStore {
	return &CoachStore{redis: redis}
}

// CachedCoachProfile is a lightweight cache representation
// Mirrors model pointer types to avoid unnecessary conversions
type CachedCoachProfile struct {
	ID                 uint               `json:"id"`
	UserID             uint               `json:"user_id"`
	BusinessName       *string            `json:"business_name,omitempty"`
	Bio                *string            `json:"bio,omitempty"`
	Specialties        []string           `json:"specialties,omitempty"`
	YearsExperience    *int               `json:"years_experience,omitempty"`
	TrainingType       string             `json:"training_type"`
	HourlyRate         *float64           `json:"hourly_rate,omitempty"`
	IsAcceptingClients bool               `json:"is_accepting_clients"`
	SocialLinks        models.SocialLinks `json:"social_links,omitempty"`
	SubscriptionTier   string             `json:"subscription_tier"`
}

// CachedCoachStats is a lightweight cache representation
type CachedCoachStats struct {
	ID                     uint `json:"id"`
	CoachID                uint `json:"coach_id"`
	ActiveClients          int  `json:"active_clients"`
	TotalClientsAllTime    int  `json:"total_clients_all_time"`
	WorkoutsAssignedTotal  int  `json:"workouts_assigned_total"`
	WorkoutsCompletedTotal int  `json:"workouts_completed_total"`
	SessionsCompletedTotal int  `json:"sessions_completed_total"`
}

// ToCachedCoachProfile converts a models.CoachProfile to cached version
func ToCachedCoachProfile(c *models.CoachProfile) *CachedCoachProfile {
	if c == nil {
		return nil
	}
	return &CachedCoachProfile{
		ID:                 c.ID,
		UserID:             c.UserID,
		BusinessName:       c.BusinessName,
		Bio:                c.Bio,
		Specialties:        c.Specialties,
		YearsExperience:    c.YearsExperience,
		TrainingType:       c.TrainingType,
		HourlyRate:         c.HourlyRate,
		IsAcceptingClients: c.IsAcceptingClients,
		SocialLinks:        c.SocialLinks,
		SubscriptionTier:   c.SubscriptionTier,
	}
}

// ToCachedCoachStats converts a models.CoachStats to cached version
func ToCachedCoachStats(s *models.CoachStats) *CachedCoachStats {
	if s == nil {
		return nil
	}
	return &CachedCoachStats{
		ID:                     s.ID,
		CoachID:                s.CoachID,
		ActiveClients:          s.ActiveClients,
		TotalClientsAllTime:    s.TotalClientsAllTime,
		WorkoutsAssignedTotal:  s.WorkoutsAssignedTotal,
		WorkoutsCompletedTotal: s.WorkoutsCompletedTotal,
		SessionsCompletedTotal: s.SessionsCompletedTotal,
	}
}

// GetProfile retrieves a cached coach profile
func (s *CoachStore) GetProfile(coachID uint) (*CachedCoachProfile, bool) {
	if !s.redis.IsAvailable() {
		return nil, false
	}

	var profile CachedCoachProfile
	if s.redis.GetJSON(KeyCoachProfile(coachID), &profile) {
		return &profile, true
	}
	return nil, false
}

// SetProfile caches a coach profile
func (s *CoachStore) SetProfile(profile *models.CoachProfile) {
	if !s.redis.IsAvailable() || profile == nil {
		return
	}

	cached := ToCachedCoachProfile(profile)
	s.redis.SetJSON(KeyCoachProfile(profile.ID), cached, CoachProfileTTL)
}

// GetStats retrieves cached coach stats
func (s *CoachStore) GetStats(coachID uint) (*CachedCoachStats, bool) {
	if !s.redis.IsAvailable() {
		return nil, false
	}

	var stats CachedCoachStats
	if s.redis.GetJSON(KeyCoachStats(coachID), &stats) {
		return &stats, true
	}
	return nil, false
}

// SetStats caches coach stats
func (s *CoachStore) SetStats(stats *models.CoachStats) {
	if !s.redis.IsAvailable() || stats == nil {
		return
	}

	cached := ToCachedCoachStats(stats)
	s.redis.SetJSON(KeyCoachStats(stats.CoachID), cached, CoachStatsTTL)
}

// InvalidateProfile removes a coach profile from cache
func (s *CoachStore) InvalidateProfile(coachID uint) {
	if s.redis.IsAvailable() {
		s.redis.Delete(KeyCoachProfile(coachID))
	}
}

// InvalidateStats removes coach stats from cache
func (s *CoachStore) InvalidateStats(coachID uint) {
	if s.redis.IsAvailable() {
		s.redis.Delete(KeyCoachStats(coachID))
	}
}

// InvalidateAll removes all cache for a coach
func (s *CoachStore) InvalidateAll(coachID uint) {
	s.InvalidateProfile(coachID)
	s.InvalidateStats(coachID)
	if s.redis.IsAvailable() {
		s.redis.Delete(KeyCoachAvailability(coachID))
	}
}
