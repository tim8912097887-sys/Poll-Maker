package middlewares

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/response"
)

func ErrorHandlerMiddleware() func (c fiber.Ctx, err error) {
	return func (c fiber.Ctx, err error) {
		switch {
		case errors.Is(err,shared.ErrAlreadyVoted):
			c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("AREADY_VOTED", err.Error(),nil))
			return
		case errors.Is(err,shared.ErrInvalidOption):
			c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("INVALID_OPTION", err.Error(),nil))
			return
		case errors.Is(err,shared.ErrPollNotFound):
			c.Status(fiber.StatusNotFound).JSON(response.NewErrorResponse("POLL_NOT_FOUND", err.Error(),nil))
			return
		default:
			c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("INTERNAL_ERROR", "internal server error",nil))
			return
		}
	}
}