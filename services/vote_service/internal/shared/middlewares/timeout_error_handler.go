package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared"
)

func TimoutErrorHandler(c fiber.Ctx) error {
	ErrorHandlerMiddleware(c, shared.ErrTimeout)
	return nil
}