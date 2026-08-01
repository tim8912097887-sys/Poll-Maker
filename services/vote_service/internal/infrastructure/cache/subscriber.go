package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
)

type VoteCache interface {
	HasVoted(ctx context.Context, pollID, sessionID string) (bool, error)
	MarkVoted(ctx context.Context, pollID, sessionID string, expiredAt time.Time) error
    GetPollMeta(ctx context.Context, pollID string) (*types.PollMeta, error)
    IsValidOption(ctx context.Context, pollID, optionID string) (bool, error)
    DeleteVoteCache(ctx context.Context, pollID string) error
}

type Subscriber struct {
	cacheClient *redis.Client
	voteCache  VoteCache
	logger    *slog.Logger
}

type SubscriberConfig struct {
	CacheClient *redis.Client
	VoteCache  VoteCache
	Logger    *slog.Logger
}

func NewSubscriber(config SubscriberConfig) *Subscriber {
	return &Subscriber{
		cacheClient: config.CacheClient,
		voteCache:  config.VoteCache,
		logger:    config.Logger,
	}
}

func (s *Subscriber) Start(ctx context.Context) {
	sub := s.cacheClient.Subscribe(ctx, DeleteCacheEvent)
    defer sub.Close()

	ch := sub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg,ok := <-ch:

			// Handle the case where the channel is closed
			if !ok {
				s.logger.Warn("Redis pub/sub channel closed")
				return
			}
			var event types.PollDeletedEvent
			err := json.Unmarshal([]byte(msg.Payload), &event)
			if err != nil {
				s.logger.Error("Failed to unmarshal poll deleted event", slog.Any("error", err))
				continue
			}
			err = s.voteCache.DeleteVoteCache(ctx, event.PollID)
			if err != nil {
				s.logger.Error("Failed to delete vote cache", slog.Any("pollID", event.PollID), slog.Any("error", err))
				continue
			}
		}
	}
}