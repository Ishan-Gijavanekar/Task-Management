package http

import (
	"strconv"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"github.com/Ishan-Gijavanekar/task-service/internal/service"
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) registerAdministration(r fiber.Router) {
	r.Post("/workspaces", h.createWorkspace)
	r.Get("/workspaces/:workspaceId", h.getWorkspace)
	r.Patch("/workspaces/:workspaceId", h.updateWorkspace)
	r.Delete("/workspaces/:workspaceId", h.archiveWorkspace)
	r.Post("/workspaces/:workspaceId/transfer", h.transferWorkspace)
	r.Post("/workspaces/:workspaceId/projects", h.createProject)
	r.Get("/workspaces/:workspaceId/projects", h.listProjects)
	r.Get("/workspaces/:workspaceId/projects/:projectId", h.getProject)
	r.Patch("/workspaces/:workspaceId/projects/:projectId", h.updateProject)
	r.Delete("/workspaces/:workspaceId/projects/:projectId", h.archiveProject)
	r.Post("/workspaces/:workspaceId/invitations", h.invite)
	r.Post("/workspaces/:workspaceId/invitations/:invitationId/resend", h.resendInvitation)
	r.Delete("/workspaces/:workspaceId/invitations/:invitationId", h.revokeInvitation)
	r.Post("/invitations/accept", h.acceptInvitation)
	r.Get("/workspaces/:workspaceId/members", h.listMembers)
	r.Patch("/workspaces/:workspaceId/members/:userId", h.updateMember)
	r.Delete("/workspaces/:workspaceId/members/:userId", h.removeMember)
	r.Get("/workspaces/:workspaceId/audit-events", h.listAudit)
}

func (h *Handler) createWorkspace(c *fiber.Ctx) error {
	var in service.CreateWorkspaceInput
	if c.BodyParser(&in) != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	out, err := h.admin.CreateWorkspace(c.UserContext(), actor(c), in)
	if err != nil {
		return fail(c, err)
	}
	return c.Status(201).JSON(out)
}
func (h *Handler) getWorkspace(c *fiber.Ctx) error {
	out, err := h.admin.GetWorkspace(c.UserContext(), actor(c), c.Params("workspaceId"), c.QueryBool("includeArchived"))
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(out)
}
func (h *Handler) updateWorkspace(c *fiber.Ctx) error {
	var in service.UpdateWorkspaceInput
	if c.BodyParser(&in) != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	out, err := h.admin.UpdateWorkspace(c.UserContext(), actor(c), c.Params("workspaceId"), in)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(out)
}
func (h *Handler) archiveWorkspace(c *fiber.Ctx) error {
	version, err := queryVersion(c)
	if err != nil {
		return fail(c, err)
	}
	out, err := h.admin.ArchiveWorkspace(c.UserContext(), actor(c), c.Params("workspaceId"), version)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(out)
}
func (h *Handler) transferWorkspace(c *fiber.Ctx) error {
	var in struct {
		NewOwnerID string `json:"newOwnerId"`
		Version    int64  `json:"version"`
	}
	if c.BodyParser(&in) != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	out, err := h.admin.TransferOwnership(c.UserContext(), actor(c), c.Params("workspaceId"), in.NewOwnerID, in.Version)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(out)
}
func (h *Handler) createProject(c *fiber.Ctx) error {
	var in service.CreateProjectInput
	if c.BodyParser(&in) != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	out, err := h.admin.CreateProject(c.UserContext(), actor(c), c.Params("workspaceId"), in)
	if err != nil {
		return fail(c, err)
	}
	return c.Status(201).JSON(out)
}
func (h *Handler) listProjects(c *fiber.Ctx) error {
	out, err := h.admin.ListProjects(c.UserContext(), actor(c), c.Params("workspaceId"), c.QueryBool("includeArchived"))
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(fiber.Map{"projects": out})
}
func (h *Handler) getProject(c *fiber.Ctx) error {
	out, err := h.admin.GetProject(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("projectId"), c.QueryBool("includeArchived"))
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(out)
}
func (h *Handler) updateProject(c *fiber.Ctx) error {
	var in service.UpdateProjectInput
	if c.BodyParser(&in) != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	out, err := h.admin.UpdateProject(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("projectId"), in)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(out)
}
func (h *Handler) archiveProject(c *fiber.Ctx) error {
	version, err := queryVersion(c)
	if err != nil {
		return fail(c, err)
	}
	out, err := h.admin.ArchiveProject(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("projectId"), version)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(out)
}
func (h *Handler) invite(c *fiber.Ctx) error {
	var in service.InviteInput
	if c.BodyParser(&in) != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	out, err := h.admin.Invite(c.UserContext(), actor(c), c.Params("workspaceId"), in)
	if err != nil {
		return fail(c, err)
	}
	return c.Status(201).JSON(out)
}
func (h *Handler) resendInvitation(c *fiber.Ctx) error {
	out, err := h.admin.ResendInvitation(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("invitationId"))
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(out)
}
func (h *Handler) revokeInvitation(c *fiber.Ctx) error {
	err := h.admin.RevokeInvitation(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("invitationId"))
	if err != nil {
		return fail(c, err)
	}
	return c.SendStatus(204)
}
func (h *Handler) acceptInvitation(c *fiber.Ctx) error {
	var in struct {
		Token string `json:"token"`
	}
	if c.BodyParser(&in) != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	out, err := h.admin.AcceptInvitation(c.UserContext(), actor(c), in.Token)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(out)
}
func (h *Handler) listMembers(c *fiber.Ctx) error {
	out, err := h.admin.ListMembers(c.UserContext(), actor(c), c.Params("workspaceId"))
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(fiber.Map{"members": out})
}
func (h *Handler) updateMember(c *fiber.Ctx) error {
	var in service.UpdateMemberInput
	if c.BodyParser(&in) != nil {
		return fail(c, domain.ErrInvalidArgument)
	}
	out, err := h.admin.UpdateMember(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("userId"), in)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(out)
}
func (h *Handler) removeMember(c *fiber.Ctx) error {
	err := h.admin.RemoveMember(c.UserContext(), actor(c), c.Params("workspaceId"), c.Params("userId"))
	if err != nil {
		return fail(c, err)
	}
	return c.SendStatus(204)
}
func (h *Handler) listAudit(c *fiber.Ctx) error {
	out, err := h.admin.ListAudit(c.UserContext(), actor(c), c.Params("workspaceId"), int64(c.QueryInt("limit", 100)))
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(fiber.Map{"events": out})
}
func queryVersion(c *fiber.Ctx) (int64, error) {
	v, err := strconv.ParseInt(c.Query("version"), 10, 64)
	if err != nil || v < 1 {
		return 0, domain.ErrInvalidArgument
	}
	return v, nil
}
