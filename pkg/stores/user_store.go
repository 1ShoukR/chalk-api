package stores

import (
	"chalk-api/pkg/models"
	"time"
)

// UserStore handles user profile caching
type UserStore struct {
	redis *RedisClient
}

// Cache TTLs for user data
const (
	UserProfileTTL = 10 * time.Minute
	UserEmailTTL   = 10 * time.Minute
)

// NewUserStore creates a new user store
func NewUserStore(redis *RedisClient) *UserStore {
	return &UserStore{redis: redis}
}

// CachedUser is a lightweight cache representation of User
// Excludes sensitive data and relations that should be loaded fresh
// Mirrors model pointer types to avoid unnecessary conversions
type CachedUser struct {
	ID            uint       `json:"id"`
	Email         string     `json:"email"`
	FirstName     string     `json:"first_name,omitempty"`
	LastName      string     `json:"last_name,omitempty"`
	Phone         *string    `json:"phone,omitempty"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	Timezone      string     `json:"timezone,omitempty"`
	IsActive      bool       `json:"is_active"`
	IsBanned      bool       `json:"is_banned"`
	EmailVerified bool       `json:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
}

// ToCachedUser converts a models.User to CachedUser
// Requires User.Profile to be preloaded for name/phone/avatar
func ToCachedUser(u *models.User) *CachedUser {
	if u == nil {
		return nil
	}
	
	cached := &CachedUser{
		ID:            u.ID,
		Email:         u.Email,
		IsActive:      u.IsActive,
		IsBanned:      u.IsBanned,
		EmailVerified: u.EmailVerified,
		LastLoginAt:   u.LastLoginAt,
	}
	
	// Include profile data if preloaded
	if u.Profile != nil {
		cached.FirstName = u.Profile.FirstName
		cached.LastName = u.Profile.LastName
		cached.Phone = u.Profile.Phone
		cached.AvatarURL = u.Profile.AvatarURL
		cached.Timezone = u.Profile.Timezone
	}
	
	return cached
}

// GetByID retrieves a cached user by ID
func (s *UserStore) GetByID(userID uint) (*CachedUser, bool) {
	if !s.redis.IsAvailable() {
		return nil, false
	}

	var user CachedUser
	if s.redis.GetJSON(KeyUserProfile(userID), &user) {
		return &user, true
	}
	return nil, false
}

// Set caches a user profile
func (s *UserStore) Set(user *models.User) {
	if !s.redis.IsAvailable() || user == nil {
		return
	}

	cached := ToCachedUser(user)
	s.redis.SetJSON(KeyUserProfile(user.ID), cached, UserProfileTTL)

	// Also cache email -> user ID mapping for quick lookups
	s.redis.Set(KeyUserByEmail(user.Email), formatUserID(user.ID), UserEmailTTL)
}

// GetIDByEmail retrieves a user ID by email (for quick existence checks)
func (s *UserStore) GetIDByEmail(email string) (uint, bool) {
	if !s.redis.IsAvailable() {
		return 0, false
	}

	val, ok := s.redis.Get(KeyUserByEmail(email))
	if !ok {
		return 0, false
	}

	var id uint
	if _, err := parseUint(val, &id); err != nil {
		return 0, false
	}
	return id, true
}

// Invalidate removes a user from cache
func (s *UserStore) Invalidate(userID uint) {
	if s.redis.IsAvailable() {
		s.redis.Delete(KeyUserProfile(userID))
	}
}

// InvalidateByEmail removes user cache by email
func (s *UserStore) InvalidateByEmail(email string) {
	if s.redis.IsAvailable() {
		// Get user ID first to invalidate profile cache too
		if id, ok := s.GetIDByEmail(email); ok {
			s.redis.Delete(KeyUserProfile(id))
		}
		s.redis.Delete(KeyUserByEmail(email))
	}
}

// helpers
func formatUserID(id uint) string {
	return formatUintString(id)
}

func formatUintString(n uint) string {
	if n == 0 {
		return "0"
	}
	// Simple conversion for reasonable IDs
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

func parseUint(val string, id *uint) (int, error) {
	var n int64
	_, err := parseCount(val, &n)
	*id = uint(n)
	return 0, err
}
