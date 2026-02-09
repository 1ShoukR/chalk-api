package stores

import (
	"chalk-api/pkg/config"
	"log/slog"
)

// StoresCollection contains all Redis-backed cache stores
type StoresCollection struct {
	// Core Redis client
	Redis *RedisClient

	// Domain-specific stores
	User         *UserStore
	Coach        *CoachStore
	Subscription *SubscriptionStore
	Exercise     *ExerciseStore
	Nutrition    *NutritionStore
	Session      *SessionStore

	// Security & rate limiting
	Security    *SecurityStore
	RateLimiter *RateLimiter
}

// InitializeStores initializes all Redis-backed stores
// If Redis is unavailable, stores will operate in fail-open mode
func InitializeStores(cfg config.Environment) (*StoresCollection, error) {
	// Initialize Redis client
	redis, err := NewRedisClient(cfg.RedisURL)
	if err != nil {
		// Log warning but don't fail - we operate in fail-open mode
		slog.Warn("Redis initialization failed, caching disabled", "error", err)
	}

	// Create all stores with the Redis client
	// Each store handles nil/unavailable Redis gracefully
	stores := &StoresCollection{
		Redis: redis,

		// Domain stores
		User:         NewUserStore(redis),
		Coach:        NewCoachStore(redis),
		Subscription: NewSubscriptionStore(redis),
		Exercise:     NewExerciseStore(redis),
		Nutrition:    NewNutritionStore(redis),
		Session:      NewSessionStore(redis),

		// Security
		Security:    NewSecurityStore(redis),
		RateLimiter: NewRateLimiter(redis),
	}

	if redis.IsAvailable() {
		slog.Info("Stores initialized with Redis caching enabled")
	} else {
		slog.Info("Stores initialized in pass-through mode (no caching)")
	}

	return stores, nil
}

// Close closes the Redis connection
func (s *StoresCollection) Close() error {
	if s.Redis != nil {
		return s.Redis.Close()
	}
	return nil
}

// IsRedisAvailable returns true if Redis caching is enabled
func (s *StoresCollection) IsRedisAvailable() bool {
	return s.Redis != nil && s.Redis.IsAvailable()
}
