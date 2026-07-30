package vote

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
)

type VoteRepository interface {
	CreateVote(ctx context.Context,id string, vote types.CreateVoteSchema) (types.CreateVoteResponse, error)
}

type VoteCache interface {
	HasVoted(ctx context.Context, pollID, sessionID string) (bool, error)
	MarkVoted(ctx context.Context, pollID, sessionID string, expiredAt time.Time) error
    GetPollMeta(ctx context.Context, pollID string) (*types.PollMeta, error)
    IsValidOption(ctx context.Context, pollID, optionID string) (bool, error)
}

type service struct{
	voteRepository VoteRepository
	voteCache VoteCache
}

type ServiceConfig struct {
	VoteRepository VoteRepository
	VoteCache VoteCache
}

func NewService(serviceConfig *ServiceConfig) *service {
	return &service{
		voteRepository: serviceConfig.VoteRepository,
		voteCache: serviceConfig.VoteCache,
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
		return types.CreateVoteDto{}, err
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