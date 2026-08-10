package httpresponse

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  *[]fiber.Map `json:"detail"`
}

type ErrorResponse struct {
	State    string   `json:"state"`
	Data     any      `json:"data"`
	Error    Error    `json:"error"`
	MetaData MetaData `json:"metaData"`
}

func NewErrorResponse(code string, message string, detail *[]fiber.Map) ErrorResponse {
	return ErrorResponse{State: "error", Data: nil, Error: Error{Code: code, Message: message, Detail: detail}, MetaData: MetaData{Timestamp: time.Now()}}
}