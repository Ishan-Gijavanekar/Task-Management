package http

import (
	"context"
	httpmiddleware "github.com/Ishan-Gijavanekar/user-service/internal/transport/http/middleware"
	"github.com/gofiber/fiber/v2"
	"time"
)

func health(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).
		JSON(fiber.Map{
			"status": "ok",
		})
}

type ReadinessChecker interface {
	Ping(context.Context) error
}

func readiness(checker ReadinessChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if checker == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "not_ready"})
		}
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := checker.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "not_ready"})
		}
		return c.JSON(fiber.Map{"status": "ready"})
	}
}

func registerRoutes(app *fiber.App, authHandler *AuthHandler, userHandler *UserHandler, authMiddleware *httpmiddleware.AuthMiddleware, checker ReadinessChecker) {
	app.Get("/health", health)
	app.Get("/health/live", health)
	app.Get("/health/ready", readiness(checker))

	api := app.Group("/api")
	v1 := api.Group("/v1")

	authGroup := v1.Group("/auth", httpmiddleware.AuthRateLimiter())

	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", authHandler.Logout)
	authGroup.Post("/forgot-password", authHandler.ForgotPassword)
	authGroup.Post("/reset-password", authHandler.ResetPassword)
	authGroup.Post("/verify-email", authHandler.VerifyEmail)
	authGroup.Post("/resend-verification", authHandler.ResendVerification)

	userGroup := v1.Group("/users", authMiddleware.Authenticate)

	userGroup.Get("/", userHandler.ListUsers)
	userGroup.Get("/:id", userHandler.GetUser)
	userGroup.Put("/:id", userHandler.UpdateUsers)
	userGroup.Delete("/:id", userHandler.DeleteUser)
}
