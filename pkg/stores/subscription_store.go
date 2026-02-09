package stores

import (
	"chalk-api/pkg/models"
	"time"
)

// SubscriptionStore handles subscription status caching
// Used for quick feature gating without hitting the database
type SubscriptionStore struct {
	redis *RedisClient
}

const (
	SubscriptionTTL = 30 * time.Minute
)

// NewSubscriptionStore creates a new subscription store
func NewSubscriptionStore(redis *RedisClient) *SubscriptionStore {
	return &SubscriptionStore{redis: redis}
}

// CachedSubscription is a lightweight representation for feature gating
// Mirrors model pointer types to avoid unnecessary conversions
type CachedSubscription struct {
	ID               uint       `json:"id"`
	UserID           uint       `json:"user_id"`
	Status           string     `json:"status"`
	ProductID        *string    `json:"product_id,omitempty"`
	Platform         *string    `json:"platform,omitempty"`
	CurrentPeriodEnd *time.Time `json:"current_period_end,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	WillRenew        bool       `json:"will_renew"`
}

// ToCachedSubscription converts a models.Subscription to cached version
func ToCachedSubscription(s *models.Subscription) *CachedSubscription {
	if s == nil {
		return nil
	}
	return &CachedSubscription{
		ID:               s.ID,
		UserID:           s.UserID,
		Status:           s.Status,
		ProductID:        s.ProductID,
		Platform:         s.Platform,
		CurrentPeriodEnd: s.CurrentPeriodEnd,
		ExpiresAt:        s.ExpiresAt,
		WillRenew:        s.WillRenew,
	}
}

// Get retrieves a cached subscription for a user
func (s *SubscriptionStore) Get(userID uint) (*CachedSubscription, bool) {
	if !s.redis.IsAvailable() {
		return nil, false
	}

	var sub CachedSubscription
	if s.redis.GetJSON(KeySubscription(userID), &sub) {
		return &sub, true
	}
	return nil, false
}

// Set caches a subscription
func (s *SubscriptionStore) Set(sub *models.Subscription) {
	if !s.redis.IsAvailable() || sub == nil {
		return
	}

	cached := ToCachedSubscription(sub)
	s.redis.SetJSON(KeySubscription(sub.UserID), cached, SubscriptionTTL)
}

// IsActive quickly checks if a user has an active subscription
// Returns true if subscription exists and is active, false otherwise
// Fail-open: returns false if Redis unavailable (forces DB check)
func (s *SubscriptionStore) IsActive(userID uint) (bool, bool) {
	cached, ok := s.Get(userID)
	if !ok {
		return false, false // cache miss, caller should check DB
	}

	// Check status and expiration
	isActive := cached.Status == "active" && 
		(cached.ExpiresAt == nil || cached.ExpiresAt.After(time.Now()))
	
	return isActive, true // cache hit
}

// Invalidate removes a subscription from cache
func (s *SubscriptionStore) Invalidate(userID uint) {
	if s.redis.IsAvailable() {
		s.redis.Delete(KeySubscription(userID))
	}
}
