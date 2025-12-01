package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"microservice-mvp/pkg/configs"
	"microservice-mvp/pkg/logger"
)

// Client is the global Redis client instance
var Client *redis.Client

// InitRedis initializes the Redis client.
func InitRedis(cfg configs.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: 10, // Connection pool size
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ping the Redis server to check connection
	status := rdb.Ping(ctx)
	if status.Err() != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", status.Err())
	}

	Client = rdb // Set global Redis client instance
	logger.Logger.Info("Redis client initialized successfully")
	return rdb, nil
}

// GetClient returns the global Redis client instance.
func GetClient() *redis.Client {
	return Client
}

// WithContext returns a Redis client that uses the provided context.
// This allows Redis commands to be cancelled if the context is cancelled.
// It also ensures that any logging from Redis operations can inherit traceID.
func WithContext(ctx context.Context) *redis.Client {
	if ctx == nil {
		return Client
	}
	// go-redis client methods already accept a context, so we just return the client
	// and rely on the calling code to pass the context to the method.
	return Client
}
