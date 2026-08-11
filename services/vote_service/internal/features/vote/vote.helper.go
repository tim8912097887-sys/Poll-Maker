package vote

func VoteCacheKey(pollID string) string {
	return "poll:" + pollID + ":voted"
}

func PollCacheKey(pollID string) string {
	return "poll:" + pollID + ":meta"
}

func PollOptionsCacheKey(pollID string) string {
	return "poll:" + pollID + ":options"
}