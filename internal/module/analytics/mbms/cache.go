package mbms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisCache implements ResultCache interface using Redis
type redisCache struct {
	client *redis.Client
	prefix string // Key prefix để namespace
}

// NewRedisCache tạo cache mới với Redis client
func NewRedisCache(client *redis.Client) ResultCache {
	return &redisCache{
		client: client,
		prefix: "dss:model:cache:",
	}
}

// Set lưu result với TTL
func (c *redisCache) Set(
	ctx context.Context,
	key string,
	result *ModelResult,
	ttl time.Duration,
) error {
	if c.client == nil {
		return errors.New("redis client is nil")
	}

	// Serialize result to JSON
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// Store in Redis với TTL
	fullKey := c.prefix + key
	err = c.client.Set(ctx, fullKey, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Get lấy cached result nếu còn valid
func (c *redisCache) Get(ctx context.Context, key string) (*ModelResult, error) {
	if c.client == nil {
		return nil, errors.New("redis client is nil")
	}

	fullKey := c.prefix + key
	data, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			// Cache miss - not an error
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	// Deserialize
	var result ModelResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

// Invalidate xóa cached result
func (c *redisCache) Invalidate(ctx context.Context, key string) error {
	if c.client == nil {
		return errors.New("redis client is nil")
	}

	fullKey := c.prefix + key
	err := c.client.Del(ctx, fullKey).Err()
	if err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	return nil
}

// Clear xóa tất cả cache
func (c *redisCache) Clear(ctx context.Context) error {
	if c.client == nil {
		return errors.New("redis client is nil")
	}

	// Find all keys với prefix
	pattern := c.prefix + "*"
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()

	keys := make([]string, 0)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	// Delete all keys
	if len(keys) > 0 {
		err := c.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}

// memoryCache implements ResultCache interface using in-memory map
// Useful cho testing hoặc khi không có Redis
type memoryCache struct {
	data map[string]*cacheEntry
}

type cacheEntry struct {
	result    *ModelResult
	expiresAt time.Time
}

// NewMemoryCache tạo in-memory cache
func NewMemoryCache() ResultCache {
	return &memoryCache{
		data: make(map[string]*cacheEntry),
	}
}

// Set lưu result với TTL
func (c *memoryCache) Set(
	ctx context.Context,
	key string,
	result *ModelResult,
	ttl time.Duration,
) error {
	c.data[key] = &cacheEntry{
		result:    result,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

// Get lấy cached result
func (c *memoryCache) Get(ctx context.Context, key string) (*ModelResult, error) {
	entry, exists := c.data[key]
	if !exists {
		return nil, nil // Cache miss
	}

	// Check expiration
	if time.Now().After(entry.expiresAt) {
		delete(c.data, key)
		return nil, nil // Expired
	}

	return entry.result, nil
}

// Invalidate xóa cached result
func (c *memoryCache) Invalidate(ctx context.Context, key string) error {
	delete(c.data, key)
	return nil
}

// Clear xóa tất cả cache
func (c *memoryCache) Clear(ctx context.Context) error {
	c.data = make(map[string]*cacheEntry)
	return nil
}
