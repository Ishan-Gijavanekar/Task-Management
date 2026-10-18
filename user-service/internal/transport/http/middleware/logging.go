package httpmiddleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Logging(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		logger.Info("http request completed",
			"request_id", c.Locals("requestid"),
			"method", c.Method(), "path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return err
	}
}
