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

type VoteCountUpdatedEvent struct {
	PollID string `json:"pollId"`
	OptionID string `json:"optionId"`
	VoteCount int `json:"voteCount"`
}