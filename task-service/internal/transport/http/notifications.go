package http

import (
	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) listNotifications(c *fiber.Ctx) error {
	items, err := h.service.ListNotifications(c.UserContext(), actor(c), c.QueryBool("unread"), int64(c.QueryInt("limit", 50)))
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(fiber.Map{"notifications": items})
}

func (h *Handler) markNotificationsRead(c *fiber.Ctx) error {
	if err := h.service.MarkNotificationsRead(c.UserContext(), actor(c), c.Params("notificationId")); err != nil {
		return fail(c, err)
	}
	return c.SendStatus(204)
}

func (h *Handler) getNotificationPreferences(c *fiber.Ctx) error {
	preferences, err := h.service.NotificationPreferences(c.UserContext(), actor(c))
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(preferences)
}

func (h *Handler) updateNotificationPreferences(c *fiber.Ctx) error {
	var preferences domain.NotificationPreferences
	if c.BodyParser(&preferences) != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	updated, err := h.service.UpdateNotificationPreferences(c.UserContext(), actor(c), preferences)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(updated)
}
