package tests

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/features/vote"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/features/vote/tests/helper"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/infrastructure/cache"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
)

func initializeSubscriber(t *testing.T, rdb *redis.Client) {
	t.Helper()

	handlerOpts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, handlerOpts))
	slog.SetDefault(logger)

	
	cacheSubscriber := cache.NewSubscriber(cache.SubscriberConfig{
		CacheClient: rdb,
		Logger:      logger,
		VoteCache:   vote.NewCache(vote.CacheConfig{
			CacheClient: rdb,
		}),
		RoomManager: &vote.RoomManager{},
	})
    
	go cacheSubscriber.Start(context.Background())
}

func TestDeleteVoteCacheEvent(t *testing.T) {
	pool, rdb, ctx, cleanup := helper.InitIntegrationDeps(t)
	defer cleanup()

	pollID := uuid.NewString()
	sessionID1 := uuid.NewString()
	sessionID2 := uuid.NewString()

	defer helper.CleanupPollData(t, ctx, pool, rdb, pollID)
	
	rdb.SAdd(ctx, vote.VoteCacheKey(pollID), sessionID1, sessionID2)

	initializeSubscriber(t, rdb)

	initializeTimer := time.NewTimer(time.Millisecond*500)

	// Wait for the subscriber to start and subscribe to the channel
	<-initializeTimer.C		
	event := types.PollDeletedEvent{
		PollID: pollID,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	_, err = rdb.Publish(
		ctx,
		cache.DeleteCacheEvent,
		data,
	).Result()

	if err != nil {
		t.Fatal(err)
	}
	// Wait for the subscriber to process the event
    processTimer := time.NewTimer(time.Millisecond*500)
    <-processTimer.C

	// Check if the cache has been deleted
	if rdb.SCard(ctx, vote.VoteCacheKey(pollID)).Val() != 0 {
		t.Fatal("expected vote cache to be deleted")
	}
}