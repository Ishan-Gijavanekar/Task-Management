package http

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func NewServer(
	logger *slog.Logger,
	health *HealthHandler,
) *fiber.App {

	app := fiber.New(
		fiber.Config{
			AppName:               "Notification Service",
			DisableStartupMessage: true,
			BodyLimit:             2 * 1024 * 1024,
			ReadTimeout:           10 * time.Second,
			IdleTimeout:           60 * time.Second,
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				code := fiber.StatusInternalServerError
				message := "internal server error"
				if fiberErr, ok := err.(*fiber.Error); ok {
					code = fiberErr.Code

					if code < 500 {
						message = fiberErr.Message
					}
				}
				if code >= 500 {
					logger.Error(
						"http request failed",
						"path", c.Path(),
						"error", err,
					)
				}
				return c.Status(code).JSON(
					fiber.Map{
						"error": message,
					},
				)
			},
		},
	)

	app.Use(requestid.New())
	app.Use(recover.New())

	app.Get("/livez", health.Liveliness)
	app.Get("/readyz", health.Readiness)

	return app
}
