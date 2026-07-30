package types

import "time"

type PollMeta struct {
	StartedAt time.Time
	ExpiredAt time.Time
	IsPrivate bool
}