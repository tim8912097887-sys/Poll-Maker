package httpresponse

import "time"

type MetaData struct {
	Timestamp time.Time `json:"timestamp"`
}