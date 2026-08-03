package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "github.com/tim8912097887-sys/Poll-Maker/services/vote_service/cmd/api/v1"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/infrastructure/cache"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/infrastructure/persistence"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/configs"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/shutdown"
)

func main() {

	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, handlerOpts))
	slog.SetDefault(logger)

	envConfigs := configs.InitConfigs()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	shutdownManager := shutdown.NewManager(logger)

	// Inialize the database connection pool and register the shutdown handler
	pool, err := persistence.Init(logger,ctx,envConfigs.Db.Url,shutdownManager)
	if err != nil {
		logger.Error("failed to initialize database connection",slog.Any("error", err))
		os.Exit(1)
	}

	// Initialize cache connection pool and register the shutdown handler
	rdb := cache.NewRedisClient(logger,envConfigs.Cache.Url)
	rdb, err = cache.CacheInit(ctx,logger,rdb,shutdownManager)
	if err != nil {
		logger.Error("failed to initialize cache connection",slog.Any("error", err))
		os.Exit(1)
	}

	v1Api := v1.Api{ Config: v1.ApiConfig{
		Logger: logger, 
		EnvConfigs: envConfigs,
		CacheClient: rdb,
		Db: pool,
		ShutdownManager: shutdownManager,
	} }

	if err := v1Api.Run(ctx, v1Api.Mount(ctx), 5*time.Second); err != nil {
		logger.Error("failed to run api",slog.Any("error", err))
		os.Exit(1)
	}
}
