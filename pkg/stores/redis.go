package stores

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client with common operations
// Implements fail-open pattern: if Redis is unavailable, operations return gracefully
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisClient creates a new Redis client wrapper
// Returns nil client if redisURL is empty (Redis is optional)
func NewRedisClient(redisURL string) (*RedisClient, error) {
	if redisURL == "" {
		slog.Warn("Redis URL not configured, caching disabled")
		return &RedisClient{client: nil, ctx: context.Background()}, nil
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		// Try treating it as host:port format
		opts = &redis.Options{
			Addr: redisURL,
		}
	}

	client := redis.NewClient(opts)
	ctx := context.Background()

	// Test connection but don't fail if Redis is down (fail-open)
	if err := client.Ping(ctx).Err(); err != nil {
		slog.Warn("Redis connection failed, caching disabled", "error", err)
		return &RedisClient{client: nil, ctx: ctx}, nil
	}

	slog.Info("Redis connected successfully")
	return &RedisClient{client: client, ctx: ctx}, nil
}

// IsAvailable returns true if Redis client is connected
func (r *RedisClient) IsAvailable() bool {
	return r.client != nil
}

// Get retrieves a value by key
// Returns empty string and false if key doesn't exist or Redis unavailable
func (r *RedisClient) Get(key string) (string, bool) {
	if r.client == nil {
		return "", false
	}

	val, err := r.client.Get(r.ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false
	}
	if err != nil {
		slog.Debug("Redis GET error", "key", key, "error", err)
		return "", false
	}
	return val, true
}

// GetJSON retrieves and unmarshals a JSON value
func (r *RedisClient) GetJSON(key string, dest interface{}) bool {
	val, ok := r.Get(key)
	if !ok {
		return false
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		slog.Debug("Redis JSON unmarshal error", "key", key, "error", err)
		return false
	}
	return true
}

// Set stores a value with expiration
func (r *RedisClient) Set(key string, value string, expiration time.Duration) bool {
	if r.client == nil {
		return false
	}

	if err := r.client.Set(r.ctx, key, value, expiration).Err(); err != nil {
		slog.Debug("Redis SET error", "key", key, "error", err)
		return false
	}
	return true
}

// SetJSON marshals and stores a value with expiration
func (r *RedisClient) SetJSON(key string, value interface{}, expiration time.Duration) bool {
	data, err := json.Marshal(value)
	if err != nil {
		slog.Debug("Redis JSON marshal error", "key", key, "error", err)
		return false
	}
	return r.Set(key, string(data), expiration)
}

// Delete removes a key
func (r *RedisClient) Delete(key string) bool {
	if r.client == nil {
		return false
	}

	if err := r.client.Del(r.ctx, key).Err(); err != nil {
		slog.Debug("Redis DEL error", "key", key, "error", err)
		return false
	}
	return true
}

// DeletePattern removes all keys matching a pattern
// Use with caution in production (KEYS command can be slow)
func (r *RedisClient) DeletePattern(pattern string) bool {
	if r.client == nil {
		return false
	}

	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		slog.Debug("Redis KEYS error", "pattern", pattern, "error", err)
		return false
	}

	if len(keys) == 0 {
		return true
	}

	if err := r.client.Del(r.ctx, keys...).Err(); err != nil {
		slog.Debug("Redis DEL pattern error", "pattern", pattern, "error", err)
		return false
	}
	return true
}

// Exists checks if a key exists
func (r *RedisClient) Exists(key string) bool {
	if r.client == nil {
		return false
	}

	count, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		slog.Debug("Redis EXISTS error", "key", key, "error", err)
		return false
	}
	return count > 0
}

// Incr increments a key's value and returns the new count
// Creates the key with value 1 if it doesn't exist
func (r *RedisClient) Incr(key string) (int64, bool) {
	if r.client == nil {
		return 0, false
	}

	val, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		slog.Debug("Redis INCR error", "key", key, "error", err)
		return 0, false
	}
	return val, true
}

// IncrWithExpiry increments a key and sets expiry if it's a new key
func (r *RedisClient) IncrWithExpiry(key string, expiration time.Duration) (int64, bool) {
	if r.client == nil {
		return 0, false
	}

	// Use a pipeline to atomically increment and set expiry
	pipe := r.client.Pipeline()
	incrCmd := pipe.Incr(r.ctx, key)
	pipe.Expire(r.ctx, key, expiration)

	_, err := pipe.Exec(r.ctx)
	if err != nil {
		slog.Debug("Redis INCR with expiry error", "key", key, "error", err)
		return 0, false
	}

	return incrCmd.Val(), true
}

// TTL returns the remaining time to live for a key
func (r *RedisClient) TTL(key string) (time.Duration, bool) {
	if r.client == nil {
		return 0, false
	}

	ttl, err := r.client.TTL(r.ctx, key).Result()
	if err != nil {
		slog.Debug("Redis TTL error", "key", key, "error", err)
		return 0, false
	}
	return ttl, true
}

// SetNX sets a key only if it doesn't exist (useful for locks)
func (r *RedisClient) SetNX(key string, value string, expiration time.Duration) bool {
	if r.client == nil {
		return false
	}

	set, err := r.client.SetNX(r.ctx, key, value, expiration).Result()
	if err != nil {
		slog.Debug("Redis SETNX error", "key", key, "error", err)
		return false
	}
	return set
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	if r.client == nil {
		return nil
	}
	return r.client.Close()
}
