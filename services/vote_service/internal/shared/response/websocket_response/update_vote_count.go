package websocketresponse

func NewUpdateVoteCountMessage(pollID string, optionID string, voteCount int) WSMessage {
	data := map[string]any{
		"poll_id":    pollID,
		"option_id":  optionID,
		"vote_count": voteCount,
	}
	return WSMessage{Type: "vote_count_update", Data: data}
}