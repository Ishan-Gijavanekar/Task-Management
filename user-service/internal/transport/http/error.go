package http

import (
	"errors"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"

	"github.com/gofiber/fiber/v2"
)

func writeError(
	c *fiber.Ctx,
	err error,
) error {

	statusCode := fiber.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := "internal server error"

	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		statusCode = fiber.StatusNotFound
		code = "USER_NOT_FOUND"
		message = "user not found"

	case errors.Is(err, domain.ErrUserAlreadyExsists):
		statusCode = fiber.StatusConflict
		code = "USER_ALREADY_EXISTS"
		message = "user already exists"

	case errors.Is(err, domain.ErrInvalidUserId):
		statusCode = fiber.StatusBadRequest
		code = "INVALID_USER_ID"
		message = "invalid user id"

	case errors.Is(err, domain.ErrInvalidName):
		statusCode = fiber.StatusBadRequest
		code = "INVALID_USER_NAME"
		message = "invalid user name"

	case errors.Is(err, domain.ErrInvalidEmail):
		statusCode = fiber.StatusBadRequest
		code = "INVALID_USER_EMAIL"
		message = "invalid user email"

	case errors.Is(err, domain.ErrInvalidPassword):
		statusCode = fiber.StatusBadRequest
		code = "INVALID_PASSWORD"
		message = "invalid password"

	case errors.Is(err, domain.ErrInvalidCredentials):
		statusCode = fiber.StatusUnauthorized
		code = "INVALID_CREDENTIALS"
		message = "invalid credentials"

	case errors.Is(err, domain.ErrUnauthorized):
		statusCode = fiber.StatusUnauthorized
		code = "UNAUTHORIZED"
		message = "authentication required"

	case errors.Is(err, domain.ErrForbidden):
		statusCode = fiber.StatusForbidden
		code = "FORBIDDEN"
		message = "permission denied"

	case errors.Is(err, domain.ErrInvalidToken):
		statusCode = fiber.StatusUnauthorized
		code = "INVALID_TOKEN"
		message = "invalid access token"

	case errors.Is(err, domain.ErrExpiredToken):
		statusCode = fiber.StatusUnauthorized
		code = "TOKEN_EXPIRED"
		message = "access token expired"
	}

	return c.Status(
		statusCode,
	).JSON(
		ErrorResponse{
			Error: ErrorDetail{
				Code:    code,
				Message: message,
			},
		},
	)
}
