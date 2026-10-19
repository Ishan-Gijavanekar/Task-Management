package http

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"github.com/Ishan-Gijavanekar/task-service/internal/service"
	"github.com/gofiber/fiber/v2"
)

type Handler struct{ service *service.Service }

func NewHandler(s *service.Service) *Handler { return &Handler{service: s} }
func (h *Handler) Register(r fiber.Router) {
	r.Post("/workspaces/:workspaceId/tasks", h.create)
	r.Get("/workspaces/:workspaceId/tasks", h.list)
	r.Get("/workspaces/:workspaceId/tasks/:id", h.get)
	r.Patch("/workspaces/:workspaceId/tasks/:id", h.update)
	r.Delete("/workspaces/:workspaceId/tasks/:id", h.delete)
	r.Post("/workspaces/:workspaceId/tasks/:id/restore", h.restore)
	r.Post("/workspaces/:workspaceId/tasks/:id/reorder", h.reorder)
	r.Patch("/workspaces/:workspaceId/tasks:batch", h.batch)
}
func actor(c *fiber.Ctx) string { id, _ := c.Locals("actorId").(string); return id }
func key(c *fiber.Ctx) string   { return strings.TrimSpace(c.Get("Idempotency-Key")) }
func (h *Handler) create(c *fiber.Ctx) error {
	var in service.CreateInput
	if e := c.BodyParser(&in); e != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	in.WorkspaceID = c.Params("workspaceId")
	in.IdempotencyKey = key(c)
	t, e := h.service.Create(c.UserContext(), actor(c), in)
	if e != nil {
		return fail(c, e)
	}
	return c.Status(201).JSON(t)
}
func (h *Handler) get(c *fiber.Ctx) error {
	t, e := h.service.Get(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("id"), c.QueryBool("includeDeleted"))
	if e != nil {
		return fail(c, e)
	}
	return c.JSON(t)
}
func (h *Handler) update(c *fiber.Ctx) error {
	var in service.UpdateInput
	if e := c.BodyParser(&in); e != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	in.WorkspaceID = c.Params("workspaceId")
	in.ID = c.Params("id")
	in.IdempotencyKey = key(c)
	t, e := h.service.Update(c.UserContext(), actor(c), in)
	if e != nil {
		return fail(c, e)
	}
	return c.JSON(t)
}
func (h *Handler) delete(c *fiber.Ctx) error {
	v, e := strconv.ParseInt(c.Query("version"), 10, 64)
	if e != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	if e = h.service.Delete(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("id"), v, key(c)); e != nil {
		return fail(c, e)
	}
	return c.SendStatus(204)
}
func (h *Handler) restore(c *fiber.Ctx) error {
	var in struct {
		Version int64 `json:"version"`
	}
	if e := c.BodyParser(&in); e != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	t, e := h.service.Restore(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("id"), in.Version, key(c))
	if e != nil {
		return fail(c, e)
	}
	return c.JSON(t)
}
func (h *Handler) reorder(c *fiber.Ctx) error {
	var in struct {
		Version       int64 `json:"version"`
		Before, After *float64
	}
	if e := c.BodyParser(&in); e != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	t, e := h.service.Reorder(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("id"), in.Version, in.Before, in.After, key(c))
	if e != nil {
		return fail(c, e)
	}
	return c.JSON(t)
}
func (h *Handler) batch(c *fiber.Ctx) error {
	var in struct {
		Items []service.BatchItem `json:"items"`
	}
	if e := c.BodyParser(&in); e != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	tasks, e := h.service.BatchUpdate(c.UserContext(), actor(c), c.Params("workspaceId"), key(c), in.Items)
	if e != nil {
		return fail(c, e)
	}
	return c.JSON(fiber.Map{"tasks": tasks})
}
func (h *Handler) list(c *fiber.Ctx) error {
	in := service.ListInput{WorkspaceID: c.Params("workspaceId"), ProjectID: c.Query("projectId"), AssigneeID: c.Query("assigneeId"), ReporterID: c.Query("reporterId"), LabelID: c.Query("labelId"), Query: c.Query("q"), Cursor: c.Query("cursor"), IncludeDeleted: c.QueryBool("includeDeleted")}
	in.Limit = int64(c.QueryInt("limit", 50))
	for _, v := range split(c.Query("status")) {
		in.Statuses = append(in.Statuses, domain.TaskStatus(v))
	}
	for _, v := range split(c.Query("priority")) {
		in.Priorities = append(in.Priorities, domain.Priority(v))
	}
	result, e := h.service.List(c.UserContext(), actor(c), in)
	if e != nil {
		return fail(c, e)
	}
	return c.JSON(result)
}
func split(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return strings.Split(v, ",")
}
func fail(c *fiber.Ctx, e error) error {
	status, code := 500, "internal"
	switch {
	case errors.Is(e, domain.ErrInvalidArgument), errors.Is(e, domain.ErrInvalidTransition), errors.Is(e, domain.ErrDependencyCycle):
		status, code = 400, "invalid_argument"
	case errors.Is(e, domain.ErrUnauthenticated):
		status, code = 401, "unauthenticated"
	case errors.Is(e, domain.ErrForbidden):
		status, code = 403, "forbidden"
	case errors.Is(e, domain.ErrNotFound):
		status, code = 404, "not_found"
	case errors.Is(e, domain.ErrConflict), errors.Is(e, domain.ErrIdempotencyConflict):
		status, code = 409, "conflict"
	}
	message := e.Error()
	if status == 500 {
		message = "internal server error"
	}
	return c.Status(status).JSON(fiber.Map{"error": fiber.Map{"code": code, "message": message, "requestId": c.GetRespHeader("X-Request-ID")}})
}
