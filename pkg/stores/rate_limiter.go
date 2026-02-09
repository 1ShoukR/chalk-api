package stores

import (
	"fmt"
	"time"
)

// RateLimiter provides sliding window rate limiting using Redis
type RateLimiter struct {
	redis *RedisClient
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redis *RedisClient) *RateLimiter {
	return &RateLimiter{redis: redis}
}

// RateLimitResult contains the result of a rate limit check
type RateLimitResult struct {
	Allowed   bool          // Whether the request is allowed
	Current   int64         // Current count in the window
	Limit     int64         // Maximum allowed in the window
	Remaining int64         // Remaining requests in the window
	ResetIn   time.Duration // Time until the window resets
}

// Check performs a rate limit check and increments the counter
// Returns true if the request is allowed, false if rate limited
// Fail-open: if Redis is unavailable, requests are allowed
func (rl *RateLimiter) Check(key string, limit int64, window time.Duration) RateLimitResult {
	// Fail-open: if Redis unavailable, allow request
	if !rl.redis.IsAvailable() {
		return RateLimitResult{
			Allowed:   true,
			Current:   0,
			Limit:     limit,
			Remaining: limit,
			ResetIn:   window,
		}
	}

	fullKey := KeyRateLimit(key, "")
	count, ok := rl.redis.IncrWithExpiry(fullKey, window)
	if !ok {
		// Redis error, fail-open
		return RateLimitResult{
			Allowed:   true,
			Current:   0,
			Limit:     limit,
			Remaining: limit,
			ResetIn:   window,
		}
	}

	ttl, _ := rl.redis.TTL(fullKey)

	return RateLimitResult{
		Allowed:   count <= limit,
		Current:   count,
		Limit:     limit,
		Remaining: max(0, limit-count),
		ResetIn:   ttl,
	}
}

// CheckWithoutIncrement checks the current count without incrementing
// Useful for displaying remaining limits without consuming one
func (rl *RateLimiter) CheckWithoutIncrement(key string, limit int64) RateLimitResult {
	if !rl.redis.IsAvailable() {
		return RateLimitResult{
			Allowed:   true,
			Current:   0,
			Limit:     limit,
			Remaining: limit,
		}
	}

	fullKey := KeyRateLimit(key, "")
	val, ok := rl.redis.Get(fullKey)
	if !ok {
		return RateLimitResult{
			Allowed:   true,
			Current:   0,
			Limit:     limit,
			Remaining: limit,
		}
	}

	var count int64
	if _, err := parseCount(val, &count); err != nil {
		return RateLimitResult{
			Allowed:   true,
			Current:   0,
			Limit:     limit,
			Remaining: limit,
		}
	}

	ttl, _ := rl.redis.TTL(fullKey)

	return RateLimitResult{
		Allowed:   count < limit,
		Current:   count,
		Limit:     limit,
		Remaining: max(0, limit-count),
		ResetIn:   ttl,
	}
}

// Reset clears the rate limit counter for a key
func (rl *RateLimiter) Reset(key string) bool {
	if !rl.redis.IsAvailable() {
		return false
	}
	return rl.redis.Delete(KeyRateLimit(key, ""))
}

// helper to parse count from string
func parseCount(val string, count *int64) (int, error) {
	return fmt.Sscanf(val, "%d", count)
}

// max returns the larger of two int64 values
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
