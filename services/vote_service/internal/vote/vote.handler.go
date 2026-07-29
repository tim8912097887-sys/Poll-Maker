package vote

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/response"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
)

type VoteService interface {
	CreateVote(vote *types.CreateVoteSchema) (types.CreateVoteDto, error)
}

type Handler struct {
	voteService VoteService
}

func NewHandler(voteService VoteService) *Handler {
	return &Handler{voteService: voteService}
}

func (h *Handler) RegisterRoutes(app fiber.Router) {
	app.Post("", h.CreateVote)
}

func (h *Handler) CreateVote(c fiber.Ctx) {
	vote := new(types.CreateVoteSchema)

    if err := c.Bind().Body(vote); err != nil {
        var validationErrors validator.ValidationErrors
        if errors.As(err, &validationErrors) {
            out := make([]fiber.Map, 0, len(validationErrors))
            for _, e := range validationErrors {
                // e.Field() - field name, e.Tag() - failed rule,
                // e.Param() - rule parameter, e.Value() - invalid value
                out = append(out, fiber.Map{
                    "field": e.Field(),
                    "rule":  e.Tag(),
                })
            }
            c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("validation_error", "invalid request body", &out))
            return
		}
		c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("invalid_request", "invalid request body",nil))
		return
    }
	createdVote, err := h.voteService.CreateVote(vote)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("internal_error", "internal server error",nil))
		return
	}

	data := map[string]any{
		"message": "Successfully created vote",
		"vote":    createdVote,
	}
	c.Status(fiber.StatusOK).JSON(response.NewSuccessResponse(data))
}
