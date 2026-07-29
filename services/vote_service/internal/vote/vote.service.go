package vote

import (
	"github.com/google/uuid"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CreateVote(vote *types.CreateVoteSchema) (types.CreateVoteDto, error) {
	id := uuid.New().String()
	return types.CreateVoteDto{
		Id: id,
		SessionId: vote.SessionId,
		PollId: vote.PollId,
		OptionId: vote.OptionId,
	}, nil
}