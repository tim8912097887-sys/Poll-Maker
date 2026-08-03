package vote

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/timeout"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/middlewares"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/response"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
)

type VoteService interface {
	CreateVote(ctx context.Context,vote types.CreateVoteSchema) (types.CreateVoteDto, error)
}

type Handler struct {
	voteService VoteService
	logger      *slog.Logger
}

type HandlerConfig struct {
	Logger     *slog.Logger
	VoteService VoteService
}

func NewHandler(handlerConfig *HandlerConfig) *Handler {
	return &Handler{
		voteService: handlerConfig.VoteService, 
		logger: handlerConfig.Logger,
	}
}

func (h *Handler) RegisterRoutes(app fiber.Router) {
	app.Post("", timeout.New(h.CreateVote,timeout.Config{
		Timeout: 3 * time.Second,
		OnTimeout: middlewares.TimoutErrorHandler,
	}))
}

func (h *Handler) CreateVote(c fiber.Ctx) error {
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
            return nil
		}
		c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("invalid_request", "invalid request body",nil))
		return nil
    }
	createdVote, err := h.voteService.CreateVote(c.Context(),vote)
	if err != nil {
		h.logger.Error("failed to create vote",slog.Any("error", err))
		middlewares.ErrorHandlerMiddleware(c, err)
		return nil
	}

	data := map[string]any{
		"message": "Successfully created vote",
		"vote":    createdVote,
	}
	c.Status(fiber.StatusOK).JSON(response.NewSuccessResponse(data))
	return nil
}
