package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/cmd/api"
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

	v1Api := api.Api{ Config: api.ApiConfig{Logger: logger, EnvConfigs: envConfigs} }

	if err := v1Api.Run(ctx, v1Api.Mount(), 5*time.Second, shutdownManager); err != nil {
		logger.Error("failed to run api",slog.Any("error", err))
		os.Exit(1)
	}
}
