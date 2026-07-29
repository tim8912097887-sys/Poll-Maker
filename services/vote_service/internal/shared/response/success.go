package response

import "time"

type SuccessResponse struct {
	State    string      `json:"state"`
	Error    any         `json:"error"`
	Data     any `json:"data"`
	MetaData MetaData    `json:"metaData"`
}

func NewSuccessResponse(data any) SuccessResponse {
	return SuccessResponse{State: "success", Data: data, Error: nil, MetaData: MetaData{Timestamp: time.Now()}}
}