package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type noopCache struct{}

func (noopCache) Get(_ context.Context, key string, _ interface{}) error { return fmt.Errorf("%w: %s", ErrCacheMiss, key) }
func (noopCache) Set(_ context.Context, _ string, _ interface{}, _ time.Duration) error { return nil }
func (noopCache) Delete(_ context.Context, _ string) error { return nil }
func (noopCache) Exists(_ context.Context, _ string) (bool, error) { return false, nil }

func NewNoopCache() Cache { return noopCache{} }

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return fmt.Errorf("%w: %s", ErrCacheMiss, key)
	}
	if err != nil {
		return fmt.Errorf("cache get: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("cache unmarshal: %w", err)
	}

	return nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}

	return nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache delete: %w", err)
	}

	return nil
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("cache exists: %w", err)
	}

	return n > 0, nil
}
