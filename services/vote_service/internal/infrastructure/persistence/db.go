package persistence

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/shutdown"
)

func Init(logger *slog.Logger,ctx context.Context,dbUrl string,shutdownManager *shutdown.Manager) (*pgxpool.Pool,error) {
	 // Initialize the connection pool
    pool, err := pgxpool.New(ctx, dbUrl)
    if err != nil {
       return nil,err
    }

    // Verify the connection
    if err := pool.Ping(ctx); err != nil {
        return nil,err
    }

    logger.Info("Connected to database")

	shutdownManager.Register(Close(pool))
	logger.Info("Registered shutdown handler for database")
	return pool, nil
}

func Close(pool *pgxpool.Pool) func(context.Context) error {

	return func(ctx context.Context) error {
		pool.Close()
		return nil
	}
}