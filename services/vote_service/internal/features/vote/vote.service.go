package vote

import (
	"context"
	"encoding/json"
	"errors"
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
	CreateVote(ctx context.Context,id string, vote types.CreateVoteSchema) (types.CreateVoteResponse, error)
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
}

type ServiceConfig struct {
	VoteRepository VoteRepository
	VoteCache VoteCache
	GrpcClient GrpcClient
	VoteProducer VoteProducer
}

func NewService(serviceConfig *ServiceConfig) *service {
	return &service{
		voteRepository: serviceConfig.VoteRepository,
		voteCache: serviceConfig.VoteCache,
		grpcClient: serviceConfig.GrpcClient,
		voteProducer: serviceConfig.VoteProducer,
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
		}

		// Check valid option
		valid, err := s.voteCache.IsValidOption(ctx, vote.PollId, vote.OptionId)
		if err != nil {
			return types.CreateVoteDto{}, err
		}
		if !valid {
			return types.CreateVoteDto{}, shared.ErrInvalidOption
		}

		// Create vote first in db to prevent race condition
		id := uuid.New().String()

		createdVote, err := s.voteRepository.CreateVote(ctx ,id, vote)
		if err != nil {
			return types.CreateVoteDto{}, err
		}

		// Mark voted
		err = s.voteCache.MarkVoted(ctx, vote.PollId, vote.SessionId, pollMeta.ExpiredAt)
		if err != nil {
			return types.CreateVoteDto{}, err
		}

		// Publish vote created event
		event := types.CreateVoteEvent{
			EventId: uuid.NewString(),
			PollId:  createdVote.PollId.String(),
			OptionId: createdVote.OptionId.String(),
			VotedAt: createdVote.CreatedAt.Format(time.RFC3339),
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			return types.CreateVoteDto{}, err
		}

		err = s.voteProducer.Publish(ctx, TopicVoteCreated, createdVote.OptionId.String(), eventBytes)
		if err != nil {
			return types.CreateVoteDto{}, err
		}

		return ToVoteDto(createdVote), nil
}

func ToVoteDto(vote types.CreateVoteResponse) types.CreateVoteDto {
	return types.CreateVoteDto{
		Id:        vote.Id.String(),
		SessionId: vote.SessionId,
		PollId:    vote.PollId.String(),
		OptionId:  vote.OptionId.String(),
	}
}