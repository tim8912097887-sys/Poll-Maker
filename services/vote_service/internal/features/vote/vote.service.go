package vote

import (
	"context"

	"github.com/google/uuid"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
)

type VoteRepository interface {
	CreateVote(ctx context.Context,id string, vote types.CreateVoteSchema) (types.CreateVoteResponse, error)
}

type service struct{
	voteRepository VoteRepository
}

type ServiceConfig struct {
	VoteRepository VoteRepository
}

func NewService(serviceConfig *ServiceConfig) *service {
	return &service{
		voteRepository: serviceConfig.VoteRepository,
	}
}

func (s *service) CreateVote(ctx context.Context,vote types.CreateVoteSchema) (types.CreateVoteDto, error) {
	id := uuid.New().String()

	createdVote, err := s.voteRepository.CreateVote(ctx ,id, vote)
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