package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log/slog"
	"storage_service/internal/config"
	"time"
)

type RedisCache struct {
	cfg    *config.CacheConfig
	client *redis.Client
	log    *slog.Logger
}

func NewRedisCache(cfg *config.CacheConfig, log *slog.Logger) (*RedisCache, error) {
	redisOpts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		log.Error("failed to parse Redis URL: %v", err)
		return nil, err
	}
	client := redis.NewClient(redisOpts)
	return &RedisCache{client: client}, nil
}

func (r *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return result, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for cache: %w", err)
	}

	return r.client.Set(ctx, key, data, time.Duration(ttl)*time.Second).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Close() {
	err := r.client.Close()
	if err != nil {
		r.log.Error(err.Error())
	}
}
