package http

import (
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthUserResponse struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Email     string          `json:"email"`
	Role      domain.UserRole `json:"role"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
}

type AuthResponse struct {
	User        AuthUserResponse `json:"user"`
	AccessToken string           `json:"access_token"`
	ExpiresIn   int64            `json:"expires_in"`
	TokenType   string           `json:"token_type"`
}

func AuthResponseFromResult(result *service.AuthResult) *AuthResponse {
	return &AuthResponse{
		User: AuthUserResponse{
			ID:    result.User.ID.Hex(),
			Name:  result.User.Name,
			Email: result.User.Email,
			Role:  result.User.Role,
			CreatedAt: result.User.CreatedAt.UTC().
				Format("2006-01-02T15:04:05Z07:00"),

			UpdatedAt: result.User.UpdatedAt.UTC().
				Format("2006-01-02T15:04:05Z07:00"),
		},

		AccessToken: result.AccessToken,
		ExpiresIn:   result.ExpiresIn,
		TokenType:   "Bearer",
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var request RegisterRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "inavlid request body",
			},
		})
	}

	result, err := h.authService.Register(c.UserContext(), service.RegisterInput{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(AuthResponseFromResult(result))
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var request LoginRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "inavlid request body",
			},
		})
	}

	result, err := h.authService.Login(c.UserContext(), service.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(AuthResponseFromResult(result))
}
