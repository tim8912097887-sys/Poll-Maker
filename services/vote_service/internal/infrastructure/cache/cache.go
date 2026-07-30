package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/shutdown"
)

func CacheInit(ctx context.Context,logger *slog.Logger,rdb *redis.Client,shutdownManager *shutdown.Manager) (*redis.Client,error) {
	
	// Test the connection using Ping
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to Redis",slog.String("pong", pong))

	shutdownManager.Register(CloseRedisClient(rdb))

	logger.Info("Registered shutdown handler for Redis")
	return rdb, nil
}

func NewRedisClient(logger *slog.Logger,redisURL string) *redis.Client {
    // Configure the client with production-ready settings
    opts, err := redis.ParseURL(redisURL)
    if err != nil {
        logger.Error("failed to parse redis url",slog.Any("error", err))
        return nil
    }

    // Apply custom connection pool & timeout overrides
    opts.PoolSize = 10
    opts.MinIdleConns = 5
    opts.PoolTimeout = 30 * time.Second
    opts.DialTimeout = 5 * time.Second
    opts.ReadTimeout = 3 * time.Second
    opts.WriteTimeout = 3 * time.Second
    opts.MaxRetries = 3
    opts.MinRetryBackoff = 8 * time.Millisecond
    opts.MaxRetryBackoff = 512 * time.Millisecond

    rdb := redis.NewClient(opts)

    return rdb
}

func CloseRedisClient(rdb *redis.Client) func(context.Context) error {

	return func(ctx context.Context) error {
		return rdb.Close()
	}
}