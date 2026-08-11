package helper

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func CleanupPollData(t *testing.T, ctx context.Context, pool *pgxpool.Pool, rdb *redis.Client, pollID string) {
	t.Helper()
	_, err := pool.Exec(ctx, `DELETE FROM outbox_events WHERE poll_id = $1`, pollID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `DELETE FROM votes WHERE poll_id = $1`, pollID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `DELETE FROM poll_options WHERE poll_id = $1`, pollID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `DELETE FROM polls WHERE id = $1`, pollID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rdb.FlushAll(ctx).Result()
	if err != nil {
		t.Fatal(err)
	}
}
