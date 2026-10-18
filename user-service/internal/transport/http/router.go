package http

import (
	httpmiddleware "github.com/Ishan-Gijavanekar/user-service/internal/transport/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func health(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).
		JSON(fiber.Map{
			"status": "ok",
		})
}

func registerRoutes(app *fiber.App, authHandler *AuthHandler, userHandler *UserHandler, authMiddleware *httpmiddleware.AuthMiddleware) {
	app.Get("/health", health)

	api := app.Group("/api")
	v1 := api.Group("/v1")

	authGroup := v1.Group("/auth", httpmiddleware.AuthRateLimiter())

	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)

	userGroup := v1.Group("/users", authMiddleware.Authenticate)

	userGroup.Get("/", userHandler.ListUsers)
	userGroup.Get("/:id", userHandler.GetUser)
	userGroup.Put("/:id", userHandler.UpdateUsers)
	userGroup.Delete("/:id", userHandler.DeleteUser)
}
