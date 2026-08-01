package types

import "time"

type PollMeta struct {
	StartedAt time.Time
	ExpiredAt time.Time
	IsPrivate bool
}

type PollDeletedEvent struct {
    PollID string `json:"pollId"`
}