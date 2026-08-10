package v1

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/cmd/websocket_server"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/features/vote"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/infrastructure/cache"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/configs"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/shutdown"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/validation"
	pollv1 "github.com/tim8912097887-sys/Poll-Maker/services/vote_service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ApiConfig struct {
	Logger *slog.Logger
	EnvConfigs configs.Configs
	Db *pgxpool.Pool
	CacheClient *redis.Client
	ShutdownManager *shutdown.Manager
	Producer sarama.SyncProducer
}

type Api struct {
	Config ApiConfig
}

func (a *Api) Mount(ctx context.Context) *fiber.App {
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
	grpcConn, err := grpc.NewClient(a.Config.EnvConfigs.Grpc.Addr,grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		a.Config.Logger.Error("failed to create grpc client",slog.Any("error", err))
	}
	grpcClient := pollv1.NewPollServiceClient(grpcConn)
	voteGrpcClient := vote.NewGrpcVoteService(grpcClient)
	voteProducer := vote.NewProducer(vote.ProducerConfig{
		Producer: a.Config.Producer,
		Logger: a.Config.Logger,
	})
	voteServiceConfig := vote.ServiceConfig{
		VoteRepository: voteRepository,
		VoteCache: voteCache,
		GrpcClient: voteGrpcClient,
		VoteProducer: voteProducer,
		Logger: a.Config.Logger,
	}
	voteService := vote.NewService(&voteServiceConfig)
	voteHandlerConfig := vote.HandlerConfig{
		VoteService: voteService, 
		Logger: a.Config.Logger,
	}
	voteHandler := vote.NewHandler(&voteHandlerConfig)
	voteHandler.RegisterRoutes(voteRouter)

	// Websocket
	roomManager := vote.NewRoomManager()
	subscriberConfig := cache.SubscriberConfig{
			CacheClient: a.Config.CacheClient,
			VoteCache: voteCache,
			Logger: a.Config.Logger,
			RoomManager: roomManager,
	}
	subscriber := cache.NewSubscriber(subscriberConfig)

	websocketHandlerConfig := websocket_server.WebSocketHandlerConfig{
		Logger: a.Config.Logger,
		VoteCache: voteCache,
		Rooms: roomManager,
	}

	webSocketHandler := websocket_server.NewWebSocketHandler(&websocketHandlerConfig)

    v1Router.Use("/ws", func(c fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	v1Router.Get("/ws/:pollID", websocket.New(func(c *websocket.Conn) {
		pollID := c.Params("pollID")
		a.Config.Logger.Info(
			"websocket connection attempt",
			slog.String("poll_id", pollID),
		)
		
		// Call your WebSocket logic here using c (websocket.Conn)
		webSocketHandler.HandleConnection(ctx, c, pollID)
	}))
	go func() {
		subscriber.Start(ctx)
	}()

	go func() {
		voteService.ProcessOutboxEvents(ctx)
	}()

	return app
}

func (a *Api) Run(ctx context.Context, app *fiber.App, shutdownTimeout time.Duration) error {

	// Channel to notify when the server is initialized failure
	serverErrorCh := make(chan error, 1)
	// Start the server with goroutine
	go func() {
		a.Config.Logger.Info("starting server",slog.String("address", a.Config.EnvConfigs.Api.Addr))
		if err := app.Listen(a.Config.EnvConfigs.Api.Addr); err != nil && err != http.ErrServerClosed {
			a.Config.Logger.Error("failed to start server",slog.Any("error", err))
			serverErrorCh <- err
		}
	}()

	// Register the shutdown handler
	a.Config.ShutdownManager.Register(a.Shutdown(app, shutdownTimeout))

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

func (a *Api) Shutdown(server *fiber.App, shutdownTimeout time.Duration) func(context.Context) error {
	
	return func(ctx context.Context) error {

		if err := server.Shutdown(); err != nil {
			a.Config.Logger.Error("failed to shut down the server",slog.Any("error", err))
			return err
		}

		a.Config.Logger.Info("server shut down gracefully")
		return nil
	}
}