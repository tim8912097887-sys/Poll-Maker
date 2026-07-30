package types

import "github.com/google/uuid"

type Vote struct {
	Id        string `json:"id"`
	SessionId string `json:"sessionId"`
	PollId    string `json:"pollId"`
	OptionId  string `json:"optionId"`
}

type CreateVoteSchema struct {
	SessionId string `json:"sessionId" validate:"required"`
	PollId    string `json:"pollId" validate:"required,uuid"`
	OptionId  string `json:"optionId" validate:"required,uuid"`
}

type CreateVoteResponse struct {
	Id        uuid.UUID `json:"id"`
	SessionId string    `json:"sessionId"`
	PollId    uuid.UUID    `json:"pollId"`
	OptionId  uuid.UUID    `json:"optionId"`
}

type CreateVoteDto struct {
	Id        string `json:"id"`
	SessionId string `json:"sessionId"`
	PollId    string `json:"pollId"`
	OptionId  string `json:"optionId"`
}

