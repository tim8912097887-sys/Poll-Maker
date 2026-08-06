package vote

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
)

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *repository {
	return &repository{db: db}
}

func (r *repository) CreateVote(ctx context.Context,id string, vote types.CreateVoteSchema, createVoteEvent types.CreateVoteEvent,expiredAt time.Time) (types.CreateVoteResponse, error) {
    // Transaction to ensure atomicity
	tx, err := r.db.Begin(ctx)
    if err != nil {
        return types.CreateVoteResponse{}, err
    }
    defer tx.Rollback(ctx)
	var createdVote types.CreateVoteResponse
	sqlString := `INSERT INTO votes (id, session_id, poll_id, option_id) VALUES ($1, $2, $3, $4) RETURNING id, session_id, poll_id, option_id, created_at;`
	err = tx.QueryRow(ctx, sqlString, id, vote.SessionId, vote.PollId, vote.OptionId).Scan(&createdVote.Id, &createdVote.SessionId, &createdVote.PollId, &createdVote.OptionId, &createdVote.CreatedAt)
	if err != nil {
		return types.CreateVoteResponse{}, err
	}
	
	sqlString = `INSERT INTO outbox_events (event_id, poll_id, option_id, expired_at, payload) VALUES ($1, $2, $3, $4, $5);`
	_, err = tx.Exec(ctx,sqlString,createVoteEvent.EventId, createVoteEvent.PollId, createVoteEvent.OptionId,expiredAt, createVoteEvent)
	
	if err != nil {
		return types.CreateVoteResponse{}, err
	}
	
	err = tx.Commit(ctx)
	if err != nil {
		return types.CreateVoteResponse{}, err
	}

	return createdVote, nil
}

func (r *repository) GetOutboxEvent(ctx context.Context) (types.CreateVoteEvent, error) {

	var createVoteEvent types.CreateVoteEvent
	var votedAt time.Time
	sqlString := `
	SELECT event_id, poll_id, option_id, created_at FROM outbox_events
	WHERE send_at IS NULL AND status = 'pending' AND expired_at > NOW() 
	LIMIT 1
	;`
	
	err := r.db.QueryRow(ctx, sqlString).Scan(&createVoteEvent.EventId, &createVoteEvent.PollId, &createVoteEvent.OptionId, &votedAt)
	if err != nil {
		// Turn sql.ErrNoRows into business error
		if errors.Is(err, sql.ErrNoRows) {
			return types.CreateVoteEvent{}, shared.ErrOutboxEventNotFound
		}
		return types.CreateVoteEvent{}, err
	}
	createVoteEvent.VotedAt = votedAt.Format(time.RFC3339)
	return createVoteEvent, nil
}

func (r *repository) UpdateOutboxEvent(ctx context.Context, eventId string) error {
	sqlString := `UPDATE outbox_events SET send_at = NOW(), status = 'sent' WHERE event_id = $1;`
	_, err := r.db.Exec(ctx, sqlString, eventId)
	if err != nil {
		return err
	}
	return nil
}