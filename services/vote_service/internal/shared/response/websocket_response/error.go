package websocketresponse

type WSError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewWSErrorResponse(code string, message string) WSMessage {
	return WSMessage{Type: "error", Data: WSError{Code: code, Message: message}}
}