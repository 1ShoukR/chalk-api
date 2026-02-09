package stores

import (
	"time"
)

// SessionStore handles JWT blacklisting and refresh token validation
type SessionStore struct {
	redis *RedisClient
}

// NewSessionStore creates a new session store
func NewSessionStore(redis *RedisClient) *SessionStore {
	return &SessionStore{redis: redis}
}

// --- JWT Blacklisting (for logout) ---

// BlacklistToken adds a JWT to the blacklist
// The token will be blacklisted until it would have naturally expired
func (s *SessionStore) BlacklistToken(tokenID string, expiresIn time.Duration) {
	if !s.redis.IsAvailable() || tokenID == "" {
		return
	}

	// Store with expiration matching the JWT's remaining lifetime
	// Value doesn't matter, we just check existence
	s.redis.Set(KeyJWTBlacklist(tokenID), "1", expiresIn)
}

// IsTokenBlacklisted checks if a JWT is blacklisted
// Fail-open: if Redis unavailable, token is considered valid
func (s *SessionStore) IsTokenBlacklisted(tokenID string) bool {
	if !s.redis.IsAvailable() || tokenID == "" {
		return false // fail-open
	}

	return s.redis.Exists(KeyJWTBlacklist(tokenID))
}

// --- Refresh Token Caching ---

// RefreshTokenData represents cached refresh token info
type RefreshTokenData struct {
	UserID    uint      `json:"user_id"`
	DeviceID  string    `json:"device_id,omitempty"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SetRefreshToken caches refresh token data for quick validation
// tokenHash should be a hash of the actual token, not the token itself
func (s *SessionStore) SetRefreshToken(tokenHash string, data RefreshTokenData, expiresIn time.Duration) {
	if !s.redis.IsAvailable() || tokenHash == "" {
		return
	}

	s.redis.SetJSON(KeyRefreshToken(tokenHash), data, expiresIn)
}

// GetRefreshToken retrieves cached refresh token data
func (s *SessionStore) GetRefreshToken(tokenHash string) (*RefreshTokenData, bool) {
	if !s.redis.IsAvailable() || tokenHash == "" {
		return nil, false
	}

	var data RefreshTokenData
	if s.redis.GetJSON(KeyRefreshToken(tokenHash), &data) {
		return &data, true
	}
	return nil, false
}

// InvalidateRefreshToken removes a refresh token from cache
func (s *SessionStore) InvalidateRefreshToken(tokenHash string) {
	if s.redis.IsAvailable() && tokenHash != "" {
		s.redis.Delete(KeyRefreshToken(tokenHash))
	}
}

// InvalidateAllUserTokens invalidates all refresh tokens for a user
// This is a pattern-based delete, use sparingly
func (s *SessionStore) InvalidateAllUserTokens(userID uint) {
	// Note: This requires knowing all token hashes for the user
	// In practice, you'd store user->tokens mapping or iterate DB
	// For now, this is a no-op placeholder - implement when needed
	_ = userID
}
