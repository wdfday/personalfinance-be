package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// DSS cache TTL: 1 hour
	dssCacheTTL = 1 * time.Hour

	// Redis key prefixes
	dssKeyPrefix = "dss:workflow"
)

// DSSCache handles caching of DSS preview results in Redis
type DSSCache struct {
	client *redis.Client
	logger *zap.Logger
}

// NewDSSCache creates a new DSS cache service
func NewDSSCache(client *redis.Client, logger *zap.Logger) *DSSCache {
	return &DSSCache{
		client: client,
		logger: logger,
	}
}

// buildKey constructs Redis key for DSS workflow data
// Format: dss:workflow:{monthID}:{userID}:{step}
func (c *DSSCache) buildKey(monthID, userID uuid.UUID, step string) string {
	return fmt.Sprintf("%s:%s:%s:%s", dssKeyPrefix, monthID.String(), userID.String(), step)
}

// SetPreview caches a DSS preview result
func (c *DSSCache) SetPreview(ctx context.Context, monthID, userID uuid.UUID, step string, data interface{}) error {
	if c.client == nil {
		c.logger.Debug("Redis unavailable, skipping cache")
		return nil
	}

	key := c.buildKey(monthID, userID, step)

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal preview data: %w", err)
	}

	if err := c.client.Set(ctx, key, bytes, dssCacheTTL).Err(); err != nil {
		c.logger.Error("Failed to cache preview",
			zap.String("key", key),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cached DSS preview",
		zap.String("key", key),
		zap.String("step", step),
		zap.Duration("ttl", dssCacheTTL))

	return nil
}

// GetPreview retrieves a cached DSS preview result
func (c *DSSCache) GetPreview(ctx context.Context, monthID, userID uuid.UUID, step string, dest interface{}) error {
	if c.client == nil {
		return fmt.Errorf("redis unavailable")
	}

	key := c.buildKey(monthID, userID, step)

	bytes, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return fmt.Errorf("preview not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get preview: %w", err)
	}

	if err := json.Unmarshal(bytes, dest); err != nil {
		return fmt.Errorf("failed to unmarshal preview: %w", err)
	}

	return nil
}

// GetAllPreviews retrieves all cached previews for a month (for FinalizeDSS)
func (c *DSSCache) GetAllPreviews(ctx context.Context, monthID, userID uuid.UUID) (map[string]interface{}, error) {
	if c.client == nil {
		return nil, fmt.Errorf("redis unavailable")
	}

	results := make(map[string]interface{})
	steps := []string{"auto_scoring", "goal_prioritization", "debt_strategy", "tradeoff"}

	for _, step := range steps {
		key := c.buildKey(monthID, userID, step)
		bytes, err := c.client.Get(ctx, key).Bytes()
		if err == redis.Nil {
			// Step was skipped or not run
			continue
		}
		if err != nil {
			c.logger.Warn("Failed to get preview",
				zap.String("step", step),
				zap.Error(err))
			continue
		}

		var data interface{}
		if err := json.Unmarshal(bytes, &data); err != nil {
			c.logger.Warn("Failed to unmarshal preview",
				zap.String("step", step),
				zap.Error(err))
			continue
		}

		results[step] = data
	}

	return results, nil
}

// ClearPreviews deletes all cached previews for a month
func (c *DSSCache) ClearPreviews(ctx context.Context, monthID, userID uuid.UUID) error {
	if c.client == nil {
		return nil
	}

	steps := []string{"auto_scoring", "goal_prioritization", "debt_strategy", "tradeoff"}
	keys := make([]string, 0, len(steps))

	for _, step := range steps {
		keys = append(keys, c.buildKey(monthID, userID, step))
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		c.logger.Error("Failed to clear previews",
			zap.Strings("keys", keys),
			zap.Error(err))
		return err
	}

	c.logger.Info("Cleared DSS previews",
		zap.String("month_id", monthID.String()),
		zap.String("user_id", userID.String()))

	return nil
}

// HasPreview checks if a preview exists in cache
func (c *DSSCache) HasPreview(ctx context.Context, monthID, userID uuid.UUID, step string) bool {
	if c.client == nil {
		return false
	}

	key := c.buildKey(monthID, userID, step)
	exists, err := c.client.Exists(ctx, key).Result()
	return err == nil && exists > 0
}
