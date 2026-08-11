package websocketresponse

type WSMessage struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}
