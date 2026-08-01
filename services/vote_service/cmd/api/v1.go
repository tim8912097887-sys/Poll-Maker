package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/features/vote"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/infrastructure/cache"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/configs"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/middlewares"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/shutdown"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/validation"
)

type ApiConfig struct {
	Logger *slog.Logger
	EnvConfigs configs.Configs
	Db *pgxpool.Pool
	CacheClient *redis.Client
	ShutdownManager *shutdown.Manager
}

type Api struct {
	Config ApiConfig
}

func (a *Api) Mount(ctx context.Context) http.Handler {
	app := fiber.New(fiber.Config{
        StructValidator: &validation.StructValidator{
			Validating: validator.New(),
        },
    })
	app.Get("/health",func (c fiber.Ctx)  {
		c.Status(http.StatusOK).JSON(fiber.Map{"status": "ok"})
	})

	v1Router := app.Group("/api/v1")
	// Initialize the vote service
	voteRouter := v1Router.Group("/votes")
	voteCacheConfig := vote.CacheConfig{CacheClient: a.Config.CacheClient}
	voteCache := vote.NewCache(voteCacheConfig)
	voteRepository := vote.NewRepository(a.Config.Db)
	voteServiceConfig := vote.ServiceConfig{
		VoteRepository: voteRepository,
		VoteCache: voteCache,
	}
	voteService := vote.NewService(&voteServiceConfig)
	voteHandlerConfig := vote.HandlerConfig{
		VoteService: voteService, 
		Logger: a.Config.Logger,
		ErrorHandler: middlewares.ErrorHandlerMiddleware(),
	}
	voteHandler := vote.NewHandler(&voteHandlerConfig)
	voteHandler.RegisterRoutes(voteRouter)

	subscriberConfig := cache.SubscriberConfig{
			CacheClient: a.Config.CacheClient,
			VoteCache: voteCache,
			Logger: a.Config.Logger,
	}
	subscriber := cache.NewSubscriber(subscriberConfig)
	go func() {
		subscriber.Start(ctx)
	}()
	return adaptor.FiberApp(app)
}

func (a *Api) Run(ctx context.Context, h http.Handler, shutdownTimeout time.Duration) error {
	server := &http.Server{
		Addr:    a.Config.EnvConfigs.Api.Addr,
		Handler: h,
		ReadTimeout:       5 * time.Second,
        ReadHeaderTimeout: 2 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       120 * time.Second,
	}

	// Channel to notify when the server is initialized failure
	serverErrorCh := make(chan error, 1)
	// Start the server with goroutine
	go func() {
		a.Config.Logger.Info("starting server",slog.String("address", a.Config.EnvConfigs.Api.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Config.Logger.Error("failed to start server",slog.Any("error", err))
			serverErrorCh <- err
		}
	}()

	// Register the shutdown handler
	a.Config.ShutdownManager.Register(a.Shutdown(server, shutdownTimeout))

	select {
		case <-ctx.Done():
			a.Config.Logger.Info("shutting down the server",slog.String("reason", ctx.Err().Error()))
		case err := <-serverErrorCh:
			return err
	}

	// Start a graceful shutdown
	a.Config.ShutdownManager.Shutdown(shutdownTimeout)

	return nil

}

func (a *Api) Shutdown(server *http.Server, shutdownTimeout time.Duration) func(context.Context) error {
	
	return func(ctx context.Context) error {

		if err := server.Shutdown(ctx); err != nil {
			a.Config.Logger.Error("failed to shut down the server",slog.Any("error", err))
			if closeErr := server.Close(); closeErr != nil {
				a.Config.Logger.Error("failed to close the server",slog.Any("error", err))
				return errors.Join(err,closeErr)
			}
			return err
		}

		a.Config.Logger.Info("server shut down gracefully")
		return nil
	}
}