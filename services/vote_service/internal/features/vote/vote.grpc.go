package vote

import (
	"context"

	pollv1 "github.com/tim8912097887-sys/Poll-Maker/services/vote_service/proto"
)

type GrpcVoteService struct {
	client pollv1.PollServiceClient
}

func NewGrpcVoteService(client pollv1.PollServiceClient) *GrpcVoteService {
	return &GrpcVoteService{
		client: client,
	}
}

func (g *GrpcVoteService) ValidatePollForVoting(ctx context.Context, pollID string) (*pollv1.ValidatePollResponse, error) {
	return g.client.ValidatePollForVoting(ctx, &pollv1.ValidatePollRequest{
		PollId: pollID,
	})
}