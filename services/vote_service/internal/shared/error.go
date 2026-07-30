package shared

import "errors"

var (
	ErrAlreadyVoted = errors.New("already voted")
	ErrInvalidOption = errors.New("invalid option")
	ErrPollNotFound = errors.New("poll not found")
)