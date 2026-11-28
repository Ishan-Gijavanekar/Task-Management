package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

type DependencyCheck func(context.Context) error

type HealthHandler struct {
	mongoCheck DependencyCheck
	natsCheck  DependencyCheck
	timeout    time.Duration
}

func NewHealthHandler(mongoCheck, natsCheck DependencyCheck, timeout time.Duration) *HealthHandler {
	return &HealthHandler{
		mongoCheck: mongoCheck,
		natsCheck:  natsCheck,
		timeout:    timeout,
	}
}

func (h *HealthHandler) Liveliness(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(
		fiber.Map{
			"status": "ok",
			"srvice": "notification-service",
		},
	)
}

func (h *HealthHandler) check(parent context.Context, check DependencyCheck) string {
	ctx, cancel := context.WithTimeout(
		parent,
		h.timeout,
	)
	defer cancel()

	if err := check(ctx); err != nil {
		return "unavailable"
	}

	return "ok"
}

func (h *HealthHandler) Readiness(c *fiber.Ctx) error {
	mongoStatus := h.check(c.UserContext(), h.mongoCheck)

	natsStatus := h.check(
		c.UserContext(),
		h.natsCheck,
	)

	status := "ok"
	statusCode := http.StatusOK

	if mongoStatus != "ok" || natsStatus != "ok" {
		status = "unavailable"
		statusCode = http.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(
		fiber.Map{
			"status":  status,
			"service": "notification-service",

			"dependencies": fiber.Map{
				"mongodb": mongoStatus,
				"nats":    natsStatus,
			},
		},
	)
}
