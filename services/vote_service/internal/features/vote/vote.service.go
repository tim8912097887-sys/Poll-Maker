package vote

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
	pollv1 "github.com/tim8912097887-sys/Poll-Maker/services/vote_service/proto"
)

type VoteProducer interface {
	Publish(ctx context.Context, topic string, key string, value []byte) error
}

type GrpcClient interface {
	ValidatePollForVoting(ctx context.Context, pollID string) (*pollv1.ValidatePollResponse, error)
}

type VoteRepository interface {
	 CreateVote(ctx context.Context,id string, vote types.CreateVoteSchema, createVoteEvent types.CreateVoteEvent,expiredAt time.Time) (types.CreateVoteResponse, error)
     GetOutboxEvent(ctx context.Context) (types.CreateVoteEvent, error)
     UpdateOutboxEvent(ctx context.Context, eventId string) error
}

type VoteCache interface {
	HasVoted(ctx context.Context, pollID, sessionID string) (bool, error)
	MarkVoted(ctx context.Context, pollID, sessionID string, expiredAt time.Time) error
    GetPollMeta(ctx context.Context, pollID string) (*types.PollMeta, error)
    IsValidOption(ctx context.Context, pollID, optionID string) (bool, error)
	DeleteVoteCache(ctx context.Context, pollID string) error
}

type service struct{
	voteRepository VoteRepository
	voteCache VoteCache
	grpcClient GrpcClient
	voteProducer VoteProducer
	logger *slog.Logger
}

type ServiceConfig struct {
	VoteRepository VoteRepository
	VoteCache VoteCache
	GrpcClient GrpcClient
	VoteProducer VoteProducer
	Logger *slog.Logger
}

func NewService(serviceConfig *ServiceConfig) *service {
	return &service{
		voteRepository: serviceConfig.VoteRepository,
		voteCache: serviceConfig.VoteCache,
		grpcClient: serviceConfig.GrpcClient,
		voteProducer: serviceConfig.VoteProducer,
		logger: serviceConfig.Logger,
	}
}

func (s *service) CreateVote(ctx context.Context,vote types.CreateVoteSchema) (types.CreateVoteDto, error) {
	
	    // Check if voted
		voted, err := s.voteCache.HasVoted(ctx, vote.PollId, vote.SessionId)
		if err != nil {
			return types.CreateVoteDto{}, err
		}
		if voted {
			return types.CreateVoteDto{}, shared.ErrAlreadyVoted
		}

		var pollMeta *types.PollMeta
		// Check poll meta
		pollMeta, err = s.voteCache.GetPollMeta(ctx, vote.PollId)
		if err != nil {
			// Fallback to grpc if cache is not available
				if errors.Is(err, shared.ErrPollNotFound) {

					validatePollResponse, err := s.grpcClient.ValidatePollForVoting(ctx, vote.PollId)
					if err != nil {
						return types.CreateVoteDto{}, err
					}
					if !validatePollResponse.IsValid {
						switch validatePollResponse.Reason {
						case pollv1.ValidatePollResponse_POLL_NOT_FOUND:
							return types.CreateVoteDto{}, shared.ErrPollNotFound
						case pollv1.ValidatePollResponse_POLL_EXPIRED:
							return types.CreateVoteDto{}, shared.ErrPollExpired
						case pollv1.ValidatePollResponse_POLL_NOT_STARTED:
							return types.CreateVoteDto{}, shared.ErrPollNotStarted
						case pollv1.ValidatePollResponse_POLL_CLOSED:
							return types.CreateVoteDto{}, shared.ErrPollClosed
						case pollv1.ValidatePollResponse_REASON_UNSPECIFIED:
							return types.CreateVoteDto{}, shared.ErrPollNotFound
						default:
							return types.CreateVoteDto{}, shared.ErrPollNotFound
						}
					} else {
						pollMeta = &types.PollMeta{
							ExpiredAt: validatePollResponse.ExpiredAt.AsTime(),
						}
					}
				} else {
					return types.CreateVoteDto{}, err
				}
		} else {
            if pollMeta.ExpiredAt.Before(time.Now()) {
				return types.CreateVoteDto{}, shared.ErrPollExpired
			}
			if pollMeta.StartedAt.After(time.Now()) {
				return types.CreateVoteDto{}, shared.ErrPollNotStarted
			}
			// Check valid option
			valid, err := s.voteCache.IsValidOption(ctx, vote.PollId, vote.OptionId)
			if err != nil {
				return types.CreateVoteDto{}, err
			}
			if !valid {
				return types.CreateVoteDto{}, shared.ErrInvalidOption
			}
		}

		// Create vote first in db to prevent race condition
		id := uuid.New().String()

		event := types.CreateVoteEvent{
			EventId: uuid.NewString(),
			PollId:  vote.PollId,
			OptionId: vote.OptionId,
			VotedAt: time.Now().Format(time.RFC3339),
		}

		createdVote, err := s.voteRepository.CreateVote(ctx ,id, vote, event, pollMeta.ExpiredAt)
		if err != nil {
			return types.CreateVoteDto{}, err
		}

		// Mark voted
		err = s.voteCache.MarkVoted(ctx, vote.PollId, vote.SessionId, pollMeta.ExpiredAt)
		if err != nil {
			return types.CreateVoteDto{}, err
		}

		return ToVoteDto(createdVote), nil
}

// Background job
func (s *service) ProcessOutboxEvents(ctx context.Context) { 
     ticker := time.NewTicker(5 *time.Second)
     defer ticker.Stop()

	 for {
		select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				outboxEvent, err := s.voteRepository.GetOutboxEvent(ctx)
				if err != nil {
					// Expected error
					if errors.Is(err, shared.ErrOutboxEventNotFound) {
						continue
					}
					s.logger.Error("failed to get outbox event", slog.Any("error", err))
					continue
				}

				// Serialize event
				eventBytes, err := json.Marshal(outboxEvent)
				if err != nil {
					s.logger.Error("failed to marshal outbox event", slog.Any("error", err))
					continue
				}
				// Process event
				err = s.voteProducer.Publish(ctx,TopicVoteCreated, outboxEvent.OptionId, eventBytes)
				// Don't update when publish failed for later retry
				if err != nil {
					s.logger.Error("failed to publish outbox event", slog.Any("error", err))
					continue
				}

				// Update outbox event
				err = s.voteRepository.UpdateOutboxEvent(ctx, outboxEvent.EventId)
				if err != nil {
					s.logger.Error("failed to update outbox event", slog.Any("error", err))
					continue
				}
		}
	 }
}

func ToVoteDto(vote types.CreateVoteResponse) types.CreateVoteDto {
	return types.CreateVoteDto{
		Id:        vote.Id.String(),
		SessionId: vote.SessionId,
		PollId:    vote.PollId.String(),
		OptionId:  vote.OptionId.String(),
	}
}