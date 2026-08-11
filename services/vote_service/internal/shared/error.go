package shared

import "errors"

var (
	ErrAlreadyVoted = errors.New("already voted")
	ErrInvalidOption = errors.New("invalid option")
	ErrPollNotFound = errors.New("poll not found")
	ErrPollExpired = errors.New("poll expired")
	ErrPollClosed = errors.New("poll closed")
	ErrPollNotStarted = errors.New("poll not started")
	ErrTimeout = errors.New("timeout")
	ErrOutboxEventNotFound = errors.New("outbox event not found")
	ErrRoomFull = errors.New("room is full")
)