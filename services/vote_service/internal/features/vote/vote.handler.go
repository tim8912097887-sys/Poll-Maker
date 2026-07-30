package vote

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/response"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
)

type VoteService interface {
	CreateVote(ctx context.Context,vote types.CreateVoteSchema) (types.CreateVoteDto, error)
}

type handler struct {
	voteService VoteService
	logger      *slog.Logger
    errorHandler    func(fiber.Ctx, error)
}

type HandlerConfig struct {
	Logger     *slog.Logger
	VoteService VoteService
	ErrorHandler    func(fiber.Ctx, error)
}

func NewHandler(handlerConfig *HandlerConfig) *handler {
	return &handler{
		voteService: handlerConfig.VoteService, 
		logger: handlerConfig.Logger,
		errorHandler: handlerConfig.ErrorHandler,
	}
}

func (h *handler) RegisterRoutes(app fiber.Router) {
	app.Post("", h.CreateVote)
}

func (h *handler) CreateVote(c fiber.Ctx) {
	var vote types.CreateVoteSchema

    if err := c.Bind().Body(&vote); err != nil {
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
	createdVote, err := h.voteService.CreateVote(c.Context(),vote)
	if err != nil {
		h.logger.Error("failed to create vote",slog.Any("error", err))
		h.errorHandler(c, err)
		return
	}

	data := map[string]any{
		"message": "Successfully created vote",
		"vote":    createdVote,
	}
	c.Status(fiber.StatusOK).JSON(response.NewSuccessResponse(data))
}
