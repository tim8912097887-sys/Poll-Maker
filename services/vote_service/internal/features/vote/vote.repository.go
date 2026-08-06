package vote

import (
	"context"
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

func (r *repository) GetOutboxEvents(ctx context.Context,limit int) ([]types.CreateVoteEvent, error) {

	events := make([]types.CreateVoteEvent, 0)
	sqlString := `
	SELECT event_id, poll_id, option_id, created_at FROM outbox_events
	WHERE send_at IS NULL AND status = 'pending' AND expired_at > NOW() 
	LIMIT $1
	;`
	
	rows, err := r.db.Query(ctx, sqlString,limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var createVoteEvent types.CreateVoteEvent
		var votedAt time.Time
		err := rows.Scan(&createVoteEvent.EventId, &createVoteEvent.PollId, &createVoteEvent.OptionId, &votedAt)
		if err != nil {
			return nil, err
		}
		createVoteEvent.VotedAt = votedAt.Format(time.RFC3339)
		events = append(events, createVoteEvent)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(events) == 0 {
        return nil, shared.ErrOutboxEventNotFound
    }

	return events, nil
}

func (r *repository) UpdateOutboxEvents(ctx context.Context, eventIds []string) error {
	sqlString := `UPDATE outbox_events SET send_at = NOW(), status = 'sent' WHERE event_id = ANY($1);`
	_, err := r.db.Exec(ctx, sqlString, eventIds)
	if err != nil {
		return err
	}
	return nil
}