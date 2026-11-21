package http

import (
	"strconv"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type commentInput struct {
	Body       string   `json:"body"`
	MentionIDs []string `json:"mentionIds"`
	Version    int64    `json:"version"`
}

func (h *Handler) createComment(c *fiber.Ctx) error {
	var in commentInput
	if c.BodyParser(&in) != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	comment, err := h.service.CreateComment(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("id"), in.Body, in.MentionIDs)
	if err != nil {
		return fail(c, err)
	}
	return c.Status(201).JSON(comment)
}

func (h *Handler) listComments(c *fiber.Ctx) error {
	page, err := h.service.ListComments(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("id"), c.Query("cursor"), int64(c.QueryInt("limit", 30)))
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(page)
}

func (h *Handler) updateComment(c *fiber.Ctx) error {
	var in commentInput
	if c.BodyParser(&in) != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	comment, err := h.service.UpdateComment(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("id"), c.Params("commentId"), in.Body, in.MentionIDs, in.Version)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(comment)
}

func (h *Handler) deleteComment(c *fiber.Ctx) error {
	version, err := strconv.ParseInt(c.Query("version"), 10, 64)
	if err != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	if err = h.service.DeleteComment(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("id"), c.Params("commentId"), version); err != nil {
		return fail(c, err)
	}
	return c.SendStatus(204)
}

func (h *Handler) listActivities(c *fiber.Ctx) error {
	page, err := h.service.ListActivities(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("id"), c.Query("cursor"), int64(c.QueryInt("limit", 30)))
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(page)
}
