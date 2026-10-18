package http

import (
	"strings"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/config"
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *service.AuthService
	cfg         config.AuthConfig
}

func NewAuthHandler(authService *service.AuthService, configs ...config.AuthConfig) *AuthHandler {
	cfg := config.AuthConfig{RefreshCookieName: "refresh_token", RefreshCookieSecure: true}
	if len(configs) > 0 {
		cfg = configs[0]
	}
	return &AuthHandler{authService: authService, cfg: cfg}
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
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Email     string            `json:"email"`
	Role      domain.UserRole   `json:"role"`
	Status    domain.UserStatus `json:"status"`
	Version   int64             `json:"version"`
	CreatedAt string            `json:"createdAt"`
	UpdatedAt string            `json:"updatedAt"`
}

type AuthResponse struct {
	User        AuthUserResponse `json:"user"`
	AccessToken string           `json:"access_token"`
	ExpiresIn   int64            `json:"expires_in"`
	TokenType   string           `json:"token_type"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
	AllSessions  bool   `json:"all_sessions"`
}

type EmailRequest struct {
	Email string `json:"email"`
}

type TokenRequest struct {
	Token string `json:"token"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func AuthResponseFromResult(result *service.AuthResult) *AuthResponse {
	return &AuthResponse{
		User: AuthUserResponse{
			ID:      result.User.ID.Hex(),
			Name:    result.User.Name,
			Email:   result.User.Email,
			Role:    result.User.Role,
			Status:  result.User.Status,
			Version: result.User.Version,
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

func (h *AuthHandler) setRefreshCookie(c *fiber.Ctx, token string, expiresIn int64) {
	if h.cfg.RefreshCookieName == "" || token == "" {
		return
	}
	c.Cookie(&fiber.Cookie{
		Name: h.cfg.RefreshCookieName, Value: token, Path: "/api/v1/auth",
		HTTPOnly: true, Secure: h.cfg.RefreshCookieSecure, SameSite: "Strict",
		Expires: time.Now().Add(time.Duration(expiresIn) * time.Second),
	})
}

func (h *AuthHandler) clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name: h.cfg.RefreshCookieName, Value: "", Path: "/api/v1/auth",
		HTTPOnly: true, Secure: h.cfg.RefreshCookieSecure, SameSite: "Strict",
		Expires: time.Unix(0, 0), MaxAge: -1,
	})
}

func (h *AuthHandler) refreshToken(c *fiber.Ctx, bodyToken string) string {
	if token := strings.TrimSpace(c.Cookies(h.cfg.RefreshCookieName)); token != "" {
		return token
	}
	return strings.TrimSpace(bodyToken)
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
	h.setRefreshCookie(c, result.RefreshToken, result.RefreshExpiresIn)

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
		Email:     request.Email,
		Password:  request.Password,
		Device:    c.Get("User-Agent"),
		IPAddress: c.IP(),
	})
	if err != nil {
		return writeError(c, err)
	}
	h.setRefreshCookie(c, result.RefreshToken, result.RefreshExpiresIn)

	return c.Status(fiber.StatusOK).JSON(AuthResponseFromResult(result))
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var request RefreshRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&request); err != nil {
			return fiber.ErrBadRequest
		}
	}
	result, err := h.authService.Refresh(c.UserContext(), service.RefreshInput{
		RefreshToken: h.refreshToken(c, request.RefreshToken),
		Device:       c.Get("User-Agent"), IPAddress: c.IP(),
	})
	if err != nil {
		return writeError(c, err)
	}
	h.setRefreshCookie(c, result.RefreshToken, result.RefreshExpiresIn)
	return c.Status(fiber.StatusOK).JSON(AuthResponseFromResult(result))
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var request LogoutRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&request); err != nil {
			return fiber.ErrBadRequest
		}
	}
	if err := h.authService.Logout(c.UserContext(), h.refreshToken(c, request.RefreshToken), request.AllSessions); err != nil {
		return writeError(c, err)
	}
	h.clearRefreshCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var request EmailRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.authService.RequestPasswordReset(c.UserContext(), request.Email); err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusAccepted).JSON(MessageResponse{Message: "if the account exists, reset instructions were sent"})
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var request ResetPasswordRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.authService.ResetPassword(c.UserContext(), request.Token, request.NewPassword); err != nil {
		return writeError(c, err)
	}
	return c.JSON(MessageResponse{Message: "password reset successfully"})
}

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var request TokenRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.authService.VerifyEmail(c.UserContext(), request.Token); err != nil {
		return writeError(c, err)
	}
	return c.JSON(MessageResponse{Message: "email verified successfully"})
}

func (h *AuthHandler) ResendVerification(c *fiber.Ctx) error {
	var request EmailRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.authService.RequestEmailVerification(c.UserContext(), request.Email); err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusAccepted).JSON(MessageResponse{Message: "if verification is required, instructions were sent"})
}
