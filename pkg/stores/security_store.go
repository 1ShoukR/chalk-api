package stores

import (
	"time"
)

// SecurityStore handles security-related caching: login attempts, password resets, etc.
// Implements fail-open: if Redis is unavailable, operations are allowed
type SecurityStore struct {
	redis *RedisClient
}

// Security limits (configurable via constants for easy tuning)
const (
	// Password reset: 3 attempts per hour
	PasswordResetLimit  = 3
	PasswordResetWindow = time.Hour

	// Magic link: 5 attempts per hour
	MagicLinkLimit  = 5
	MagicLinkWindow = time.Hour

	// Failed login tracking window (for monitoring, not lockout)
	LoginAttemptWindow = 15 * time.Minute

	// Invoice creation: 10 per hour per coach-client pair (lenient)
	InvoiceLimit  = 10
	InvoiceWindow = time.Hour
)

// NewSecurityStore creates a new security store
func NewSecurityStore(redis *RedisClient) *SecurityStore {
	return &SecurityStore{redis: redis}
}

// --- Login Attempt Tracking (no lockout, just monitoring) ---

// RecordLoginAttempt records a failed login attempt for an email
// Returns the current count of failed attempts
func (s *SecurityStore) RecordLoginAttempt(email string) int64 {
	if !s.redis.IsAvailable() {
		return 0
	}

	key := KeyLoginAttempts(email)
	count, _ := s.redis.IncrWithExpiry(key, LoginAttemptWindow)
	return count
}

// GetLoginAttempts returns the current failed login attempt count
func (s *SecurityStore) GetLoginAttempts(email string) int64 {
	if !s.redis.IsAvailable() {
		return 0
	}

	val, ok := s.redis.Get(KeyLoginAttempts(email))
	if !ok {
		return 0
	}

	var count int64
	parseCount(val, &count)
	return count
}

// ClearLoginAttempts clears failed login attempts after successful login
func (s *SecurityStore) ClearLoginAttempts(email string) {
	if s.redis.IsAvailable() {
		s.redis.Delete(KeyLoginAttempts(email))
	}
}

// --- Password Reset Rate Limiting ---

// CheckPasswordResetAllowed checks if a password reset request is allowed
// Returns true if under limit, false if rate limited
func (s *SecurityStore) CheckPasswordResetAllowed(email string) bool {
	if !s.redis.IsAvailable() {
		return true // fail-open
	}

	key := KeyPasswordResetAttempts(email)
	count, ok := s.redis.IncrWithExpiry(key, PasswordResetWindow)
	if !ok {
		return true // fail-open
	}

	return count <= PasswordResetLimit
}

// GetPasswordResetRemaining returns remaining password reset attempts
func (s *SecurityStore) GetPasswordResetRemaining(email string) int64 {
	if !s.redis.IsAvailable() {
		return PasswordResetLimit
	}

	val, ok := s.redis.Get(KeyPasswordResetAttempts(email))
	if !ok {
		return PasswordResetLimit
	}

	var count int64
	parseCount(val, &count)
	return max(0, PasswordResetLimit-count)
}

// --- Magic Link Rate Limiting ---

// CheckMagicLinkAllowed checks if a magic link request is allowed
func (s *SecurityStore) CheckMagicLinkAllowed(email string) bool {
	if !s.redis.IsAvailable() {
		return true
	}

	key := KeyMagicLinkAttempts(email)
	count, ok := s.redis.IncrWithExpiry(key, MagicLinkWindow)
	if !ok {
		return true
	}

	return count <= MagicLinkLimit
}

// --- Invoice Rate Limiting (lenient) ---

// CheckInvoiceAllowed checks if a coach can create an invoice for a client
// Lenient limit to prevent harassment but allow normal operations
func (s *SecurityStore) CheckInvoiceAllowed(coachID, clientID uint) bool {
	if !s.redis.IsAvailable() {
		return true
	}

	key := KeyRateLimit(formatCoachClient(coachID, clientID), "invoice")
	count, ok := s.redis.IncrWithExpiry(key, InvoiceWindow)
	if !ok {
		return true
	}

	return count <= InvoiceLimit
}

// GetInvoiceRemaining returns remaining invoice creations for coach-client pair
func (s *SecurityStore) GetInvoiceRemaining(coachID, clientID uint) int64 {
	if !s.redis.IsAvailable() {
		return InvoiceLimit
	}

	key := KeyRateLimit(formatCoachClient(coachID, clientID), "invoice")
	val, ok := s.redis.Get(key)
	if !ok {
		return InvoiceLimit
	}

	var count int64
	parseCount(val, &count)
	return max(0, InvoiceLimit-count)
}

// formatCoachClient creates a unique identifier for coach-client pair
func formatCoachClient(coachID, clientID uint) string {
	return formatUintSafe(coachID) + ":" + formatUintSafe(clientID)
}

func formatUintSafe(n uint) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
