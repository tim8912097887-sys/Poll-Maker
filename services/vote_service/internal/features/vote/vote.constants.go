package vote

import "time"

const (
	TopicVoteCreated = "vote.created"
	WriteWait        = 10 * time.Second
	PongWait         = 60 * time.Second
	PingPeriod       = (PongWait * 9) / 10
)