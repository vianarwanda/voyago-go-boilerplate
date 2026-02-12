package database

import (
	"context"
	"fmt"
	"voyago/core-api/internal/infrastructure/config"
	"voyago/core-api/internal/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
	log    logger.Logger
}

func NewRedisCache(cfg *config.RedisConfig, log logger.Logger) CacheDatabase {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.WithFields(map[string]any{
			"error": err.Error(),
		}).Warn("Failed to connect to Redis")
	}

	return &redisCache{
		client: client,
		log:    log,
	}
}

func (r *redisCache) GetClient() *redis.Client {
	return r.client
}

func (r *redisCache) Close() error {
	return r.client.Close()
}
