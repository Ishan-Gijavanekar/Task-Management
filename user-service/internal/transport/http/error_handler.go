package http

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

func NewErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		statusCode := fiber.StatusInternalServerError
		code := "INTERNAL_ERROR"
		message := "internal server error"

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			statusCode = fiberErr.Code
			switch statusCode {
			case fiber.StatusNotFound:
				code, message = "NOT_FOUND", "resource not found"
			case fiber.StatusMethodNotAllowed:
				code, message = "METHOD_NOT_ALLOWED", "method not allowed"
			case fiber.StatusBadRequest:
				code, message = "BAD_REQUEST", "bad request"
			case fiber.StatusRequestEntityTooLarge:
				code, message = "PAYLOAD_TOO_LARGE", "request body too large"
			}
		}

		if statusCode >= 500 {
			logger.Error("http request failed", "error", err, "path", c.Path(), "method", c.Method())
		}

		return c.Status(statusCode).JSON(ErrorResponse{Error: ErrorDetail{Code: code, Message: message}})
	}
}
