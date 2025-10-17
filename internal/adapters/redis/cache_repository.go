package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	config "assets-service/configs"
	"assets-service/internal/ports"

	"github.com/go-redis/redis/v8"
)

// RedisCacheService implements the CacheService interface using Redis
type RedisCacheService struct {
	client *redis.Client
	logger ports.Logger
}

// NewRedisCacheService creates a new Redis cache service
func NewRedisCacheService(client *redis.Client, logger ports.Logger) ports.CacheService {
	return &RedisCacheService{
		client: client,
		logger: logger,
	}
}

// NewRedisClient creates a new Redis client with the given configuration
func NewRedisClient(config config.RedisConfig) *redis.Client {
	addr := config.RedisURL()
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       config.DB,
		Password: config.Password,
	})
}

// Set stores a value in Redis cache
func (s *RedisCacheService) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	// Serialize the value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Set the value in Redis with TTL
	err = s.client.Set(ctx, key, data, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to set value in Redis: %w", err)
	}

	return nil
}

// Get retrieves a value from Redis cache
func (s *RedisCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	// Get the value from Redis
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found")
		}
		return fmt.Errorf("failed to get value from Redis: %w", err)
	}

	// Handle simple types directly for performance
	switch v := dest.(type) {
	case *string:
		// If the stored value is a simple string (not JSON), return it directly
		var str string
		if err := json.Unmarshal([]byte(data), &str); err != nil {
			// If JSON unmarshal fails, treat it as a plain string
			*v = data
		} else {
			*v = str
		}
	case *int:
		var i int
		if err := json.Unmarshal([]byte(data), &i); err != nil {
			return fmt.Errorf("failed to unmarshal int value: %w", err)
		}
		*v = i
	case *float64:
		var f float64
		if err := json.Unmarshal([]byte(data), &f); err != nil {
			return fmt.Errorf("failed to unmarshal float64 value: %w", err)
		}
		*v = f
	case *bool:
		var b bool
		if err := json.Unmarshal([]byte(data), &b); err != nil {
			return fmt.Errorf("failed to unmarshal bool value: %w", err)
		}
		*v = b
	default:
		// For complex types, unmarshal from JSON
		if err := json.Unmarshal([]byte(data), dest); err != nil {
			return fmt.Errorf("failed to unmarshal value: %w", err)
		}
	}

	return nil
}

// Delete removes a value from Redis cache
func (s *RedisCacheService) Delete(ctx context.Context, key string) error {
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key from Redis: %w", err)
	}
	return nil
}

// Ping tests the Redis connection
func (s *RedisCacheService) Ping(ctx context.Context) error {
	err := s.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("Redis ping failed: %w", err)
	}
	return nil
}

// Close closes the Redis connection
func (s *RedisCacheService) Close() error {
	return s.client.Close()
}

// FlushAll clears all data from Redis (use with caution!)
func (s *RedisCacheService) FlushAll(ctx context.Context) error {
	err := s.client.FlushAll(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush Redis: %w", err)
	}
	return nil
}

// SetWithExpiration sets a value with a specific expiration time
func (s *RedisCacheService) SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Time) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	ttl := time.Until(expiration)
	if ttl <= 0 {
		return fmt.Errorf("expiration time is in the past")
	}

	err = s.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set value in Redis: %w", err)
	}

	return nil
}

// Exists checks if a key exists in Redis
func (s *RedisCacheService) Exists(ctx context.Context, key string) (bool, error) {
	count, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return count > 0, nil
}

// GetTTL returns the remaining TTL for a key
func (s *RedisCacheService) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}
	return ttl, nil
}

// Increment increments a numeric value in Redis
func (s *RedisCacheService) Increment(ctx context.Context, key string) (int64, error) {
	val, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment value: %w", err)
	}
	return val, nil
}

// IncrementBy increments a numeric value by a specific amount
func (s *RedisCacheService) IncrementBy(ctx context.Context, key string, increment int64) (int64, error) {
	val, err := s.client.IncrBy(ctx, key, increment).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment value: %w", err)
	}
	return val, nil
}

// SetIfNotExists sets a value only if the key doesn't exist (atomic operation)
func (s *RedisCacheService) SetIfNotExists(ctx context.Context, key string, value interface{}, ttl int) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	success, err := s.client.SetNX(ctx, key, data, time.Duration(ttl)*time.Second).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set value if not exists: %w", err)
	}
	return success, nil
}
