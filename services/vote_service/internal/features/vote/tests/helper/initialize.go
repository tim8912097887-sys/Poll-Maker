package helper

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	cacheinfra "github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/infrastructure/cache"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/infrastructure/persistence"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/configs"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/shutdown"
)

func InitIntegrationDeps(t *testing.T) (*pgxpool.Pool, *redis.Client, context.Context, func()) {
	t.Helper()

	handlerOpts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, handlerOpts))
	slog.SetDefault(logger)

	envConfigs := configs.InitConfigs()
	ctx, cancel := context.WithCancel(context.Background())

	shutdownManager := shutdown.NewManager(logger)

	pool, err := persistence.Init(logger, ctx, envConfigs.Db.Url, shutdownManager)
	if err != nil {
		t.Skipf("skipping integration test because database is unavailable: %v", err)
	}

	rdb := cacheinfra.NewRedisClient(logger, envConfigs.Cache.Url)
	rdb, err = cacheinfra.CacheInit(ctx, logger, rdb, shutdownManager)
	if err != nil {
		cancel()
		t.Skipf("skipping integration test because redis is unavailable: %v", err)
	}

	return pool, rdb, ctx, func() {
		shutdownManager.Shutdown(5 * time.Second)
		cancel()
	}
}
