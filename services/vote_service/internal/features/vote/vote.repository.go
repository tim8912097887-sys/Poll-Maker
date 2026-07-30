package vote

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
)

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *repository {
	return &repository{db: db}
}

func (r *repository) CreateVote(ctx context.Context,id string, vote types.CreateVoteSchema) (types.CreateVoteResponse, error) {
    
	var createdVote types.CreateVoteResponse
	sqlString := `INSERT INTO votes (id, session_id, poll_id, option_id) VALUES ($1, $2, $3, $4) RETURNING id, session_id, poll_id, option_id;`
	err := r.db.QueryRow(ctx, sqlString, id, vote.SessionId, vote.PollId, vote.OptionId).Scan(&createdVote.Id, &createdVote.SessionId, &createdVote.PollId, &createdVote.OptionId)
	if err != nil {
		return types.CreateVoteResponse{}, err
	}
	return createdVote, nil
}