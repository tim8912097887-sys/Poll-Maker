package vote

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
)

type cache struct{
	cacheClient *redis.Client
}

type CacheConfig struct {
	CacheClient *redis.Client
}

func NewCache(config CacheConfig) *cache {
	return &cache{
		cacheClient: config.CacheClient,
	}
}

func (c *cache) HasVoted(ctx context.Context, pollID, sessionID string) (bool, error) {
	
	key := fmt.Sprintf("poll:%s:voted", pollID)

	voted, err := c.cacheClient.SIsMember(ctx, key, sessionID).Result()
	if err != nil {
		return false, err
	}
	return voted, nil
}

func (c *cache) MarkVoted(
    ctx context.Context,
    pollID,
    sessionID string,
    expiredAt time.Time,
) error {

    key := fmt.Sprintf("poll:%s:voted", pollID)

    pipe := c.cacheClient.TxPipeline()

    pipe.SAdd(ctx, key, sessionID)

    ttl := time.Until(expiredAt)
    if ttl > 0 {
        pipe.ExpireNX(ctx, key, ttl)
    }

    _, err := pipe.Exec(ctx)
    return err
}

func (c *cache) GetPollMeta(ctx context.Context, pollID string) (*types.PollMeta, error) {
    key := fmt.Sprintf("poll:%s:meta", pollID)

    res := c.cacheClient.HGetAll(ctx, key)
    if err := res.Err(); err != nil {
        return nil, err
    }

    // HGetAll returns an empty map if the key does not exist
    resultMap, err := res.Result()
    if err != nil {
        return nil, err
    }
    if len(resultMap) == 0 {
        return nil, shared.ErrPollNotFound 
    }

    var pollMeta types.PollMeta
    if err := res.Scan(&pollMeta); err != nil {
        return nil, err
    }

    return &pollMeta, nil
}
func (c *cache) IsValidOption(ctx context.Context, pollID, optionID string) (bool, error) {
	
	key := fmt.Sprintf("poll:%s:options", pollID)

	valid, err := c.cacheClient.SIsMember(ctx, key, optionID).Result()
	if err != nil {
		return false, err
	}
	return valid, nil
}