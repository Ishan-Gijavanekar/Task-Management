package httpmiddleware

import (
	"strings"

	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	JwtManager *auth.JWTManager
}

func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		JwtManager: jwtManager,
	}
}

func unauthorized(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).
		JSON(fiber.Map{
			"error": fiber.Map{
				"status":  "Unauthorized",
				"message": message,
			},
		})
}

func (a *AuthMiddleware) Authenticate(c *fiber.Ctx) error {
	authorization := c.Get("authorization")
	authorization = strings.TrimSpace(authorization)

	if authorization == "" {
		return unauthorized(c, "Bearer token with authorization is required")
	}

	parts := strings.Fields(authorization)
	if len(parts) != 2 {
		return unauthorized(c, "Breaer Token isrequired")
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		unauthorized(c, "invalid token")
	}

	claims, err := a.JwtManager.ValidateAccessToken(parts[1])
	if err != nil {
		return unauthorized(c, err.Error())
	}

	authenticateUser := auth.AuthenticatedUser{
		ID:    claims.ID,
		Email: claims.Email,
		Role:  claims.Role,
	}

	ctx := auth.WithAuthenticatedUser(c.UserContext(), &authenticateUser)
	c.SetUserContext(ctx)

	return c.Next()
}
