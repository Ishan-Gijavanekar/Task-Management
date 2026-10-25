package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"github.com/Ishan-Gijavanekar/task-service/internal/policy"
	"github.com/Ishan-Gijavanekar/task-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const invitationLimitPerHour int64 = 10

var projectKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`)

type AdministrationService struct {
	repo       repository.AdministrationRepository
	identities IdentityValidator
	authorizer *Authorizer
	now        func() time.Time
}

func NewAdministration(repo repository.AdministrationRepository, identities IdentityValidator, authorizer *Authorizer) (*AdministrationService, error) {
	if repo == nil || identities == nil || authorizer == nil {
		return nil, fmt.Errorf("repository, identity validator, and authorizer are required")
	}
	return &AdministrationService{repo: repo, identities: identities, authorizer: authorizer, now: time.Now}, nil
}

type CreateWorkspaceInput struct {
	Name string `json:"name"`
}
type UpdateWorkspaceInput struct {
	Name    string `json:"name"`
	Version int64  `json:"version"`
}
type CreateProjectInput struct {
	Name        string                   `json:"name"`
	Key         string                   `json:"key"`
	Description string                   `json:"description"`
	Workflow    []domain.WorkflowStatus  `json:"workflow"`
	Visibility  domain.ProjectVisibility `json:"visibility"`
}
type UpdateProjectInput struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Workflow    []domain.WorkflowStatus  `json:"workflow"`
	Visibility  domain.ProjectVisibility `json:"visibility"`
	Version     int64                    `json:"version"`
}
type InviteInput struct {
	Email     string               `json:"email"`
	Role      domain.WorkspaceRole `json:"role"`
	ExpiresIn time.Duration        `json:"-"`
}
type InvitationResult struct {
	Invitation *domain.Invitation `json:"invitation"`
	Token      string             `json:"token"`
}
type UpdateMemberInput struct {
	Role   domain.WorkspaceRole    `json:"role"`
	Status domain.MembershipStatus `json:"status"`
}

func (s *AdministrationService) CreateWorkspace(ctx context.Context, actor string, in CreateWorkspaceInput) (*domain.Workspace, error) {
	if err := s.identities.Validate(ctx, actor); err != nil {
		return nil, err
	}
	name, err := workspaceName(in.Name)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	w := &domain.Workspace{ID: primitive.NewObjectID(), Name: name, OwnerID: actor, Version: 1, CreatedAt: now, UpdatedAt: now}
	m := &domain.Membership{ID: primitive.NewObjectID(), WorkspaceID: w.ID, UserID: actor, Role: domain.RoleOwner, Status: domain.MembershipActive, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateWorkspace(ctx, w, m, audit(w.ID, actor, "workspace.created", "workspace", w.ID.Hex(), nil, now)); err != nil {
		return nil, err
	}
	s.authorizer.Invalidate(w.ID, actor)
	return w, nil
}

func (s *AdministrationService) GetWorkspace(ctx context.Context, actor, workspaceID string, includeArchived bool) (*domain.Workspace, error) {
	wid, err := parseID(workspaceID, "workspaceId")
	if err != nil {
		return nil, err
	}
	if _, err = s.require(ctx, wid, actor, policy.WorkspaceRead); err != nil {
		return nil, err
	}
	return s.repo.GetWorkspace(ctx, wid, includeArchived)
}

func (s *AdministrationService) UpdateWorkspace(ctx context.Context, actor, workspaceID string, in UpdateWorkspaceInput) (*domain.Workspace, error) {
	wid, err := parseID(workspaceID, "workspaceId")
	if err != nil {
		return nil, err
	}
	if _, err = s.require(ctx, wid, actor, policy.WorkspaceUpdate); err != nil {
		return nil, err
	}
	name, err := workspaceName(in.Name)
	if err != nil || in.Version < 1 {
		return nil, domain.ErrInvalidArgument
	}
	w, err := s.repo.GetWorkspace(ctx, wid, false)
	if err != nil {
		return nil, err
	}
	if w.Version != in.Version {
		return nil, domain.ErrConflict
	}
	w.Name, w.UpdatedAt, w.Version = name, s.now().UTC(), w.Version+1
	if err = s.repo.UpdateWorkspace(ctx, w, in.Version, audit(wid, actor, "workspace.updated", "workspace", wid.Hex(), nil, w.UpdatedAt)); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *AdministrationService) ArchiveWorkspace(ctx context.Context, actor, workspaceID string, version int64) (*domain.Workspace, error) {
	wid, err := parseID(workspaceID, "workspaceId")
	if err != nil {
		return nil, err
	}
	if _, err = s.require(ctx, wid, actor, policy.WorkspaceArchive); err != nil {
		return nil, err
	}
	w, err := s.repo.GetWorkspace(ctx, wid, false)
	if err != nil {
		return nil, err
	}
	if version < 1 || w.Version != version {
		return nil, domain.ErrConflict
	}
	now := s.now().UTC()
	w.ArchivedAt, w.UpdatedAt, w.Version = &now, now, w.Version+1
	if err = s.repo.UpdateWorkspace(ctx, w, version, audit(wid, actor, "workspace.archived", "workspace", wid.Hex(), nil, now)); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *AdministrationService) TransferOwnership(ctx context.Context, actor, workspaceID, nextOwnerID string, version int64) (*domain.Workspace, error) {
	wid, err := parseID(workspaceID, "workspaceId")
	if err != nil {
		return nil, err
	}
	current, err := s.require(ctx, wid, actor, policy.WorkspaceTransfer)
	if err != nil {
		return nil, err
	}
	nextOwnerID = strings.TrimSpace(nextOwnerID)
	if nextOwnerID == "" || nextOwnerID == actor || version < 1 {
		return nil, domain.ErrInvalidArgument
	}
	next, err := s.repo.GetMembership(ctx, wid, nextOwnerID)
	if err != nil {
		return nil, err
	}
	if next.Status != domain.MembershipActive {
		return nil, domain.ErrConflict
	}
	w, err := s.repo.GetWorkspace(ctx, wid, false)
	if err != nil {
		return nil, err
	}
	if w.Version != version || w.OwnerID != actor {
		return nil, domain.ErrConflict
	}
	now := s.now().UTC()
	previous := *current
	previous.Role, previous.UpdatedAt = domain.RoleAdmin, now
	replacement := *next
	replacement.Role, replacement.UpdatedAt = domain.RoleOwner, now
	w.OwnerID, w.UpdatedAt, w.Version = nextOwnerID, now, w.Version+1
	details := map[string]any{"previousOwnerId": actor, "nextOwnerId": nextOwnerID}
	if err = s.repo.TransferOwnership(ctx, w, version, &previous, &replacement, audit(wid, actor, "workspace.ownership_transferred", "workspace", wid.Hex(), details, now)); err != nil {
		return nil, err
	}
	s.authorizer.Invalidate(wid, actor, nextOwnerID)
	return w, nil
}

func defaultWorkflow() []domain.WorkflowStatus {
	return []domain.WorkflowStatus{{Key: "backlog", Name: "Backlog"}, {Key: "todo", Name: "To do"}, {Key: "in_progress", Name: "In progress"}, {Key: "done", Name: "Done"}}
}

func (s *AdministrationService) CreateProject(ctx context.Context, actor, workspaceID string, in CreateProjectInput) (*domain.Project, error) {
	wid, err := parseID(workspaceID, "workspaceId")
	if err != nil {
		return nil, err
	}
	if _, err = s.require(ctx, wid, actor, policy.ProjectWrite); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(in.Name)
	key := strings.ToUpper(strings.TrimSpace(in.Key))
	if in.Visibility == "" {
		in.Visibility = domain.ProjectWorkspace
	}
	if len(in.Workflow) == 0 {
		in.Workflow = defaultWorkflow()
	}
	if err = validateProject(name, key, in.Description, in.Workflow, in.Visibility); err != nil {
		return nil, err
	}
	now := s.now().UTC()
	p := &domain.Project{ID: primitive.NewObjectID(), WorkspaceID: wid, Name: name, Key: key, Description: strings.TrimSpace(in.Description), Workflow: in.Workflow, Visibility: in.Visibility, Version: 1, CreatedAt: now, UpdatedAt: now}
	if err = s.repo.CreateProject(ctx, p, audit(wid, actor, "project.created", "project", p.ID.Hex(), nil, now)); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *AdministrationService) GetProject(ctx context.Context, actor, workspaceID, projectID string, archived bool) (*domain.Project, error) {
	wid, pid, err := twoIDs(workspaceID, projectID)
	if err != nil {
		return nil, err
	}
	membership, err := s.require(ctx, wid, actor, policy.ProjectRead)
	if err != nil {
		return nil, err
	}
	project, err := s.repo.GetProject(ctx, wid, pid, archived)
	if err != nil {
		return nil, err
	}
	if project.Visibility == domain.ProjectPrivate && membership.Role != domain.RoleOwner && membership.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	return project, nil
}

func (s *AdministrationService) ListProjects(ctx context.Context, actor, workspaceID string, archived bool) ([]*domain.Project, error) {
	wid, err := parseID(workspaceID, "workspaceId")
	if err != nil {
		return nil, err
	}
	membership, err := s.require(ctx, wid, actor, policy.ProjectRead)
	if err != nil {
		return nil, err
	}
	projects, err := s.repo.ListProjects(ctx, wid, archived)
	if err != nil {
		return nil, err
	}
	if membership.Role == domain.RoleOwner || membership.Role == domain.RoleAdmin {
		return projects, nil
	}
	visible := projects[:0]
	for _, project := range projects {
		if project.Visibility == domain.ProjectWorkspace {
			visible = append(visible, project)
		}
	}
	return visible, nil
}

func (s *AdministrationService) UpdateProject(ctx context.Context, actor, workspaceID, projectID string, in UpdateProjectInput) (*domain.Project, error) {
	wid, pid, err := twoIDs(workspaceID, projectID)
	if err != nil {
		return nil, err
	}
	if _, err = s.require(ctx, wid, actor, policy.ProjectWrite); err != nil {
		return nil, err
	}
	p, err := s.repo.GetProject(ctx, wid, pid, false)
	if err != nil {
		return nil, err
	}
	if p.Version != in.Version {
		return nil, domain.ErrConflict
	}
	if in.Visibility == "" {
		in.Visibility = p.Visibility
	}
	if len(in.Workflow) == 0 {
		in.Workflow = p.Workflow
	}
	name := strings.TrimSpace(in.Name)
	if err = validateProject(name, p.Key, in.Description, in.Workflow, in.Visibility); err != nil {
		return nil, err
	}
	p.Name, p.Description, p.Workflow, p.Visibility = name, strings.TrimSpace(in.Description), in.Workflow, in.Visibility
	p.Version++
	p.UpdatedAt = s.now().UTC()
	if err = s.repo.UpdateProject(ctx, p, in.Version, audit(wid, actor, "project.updated", "project", p.ID.Hex(), nil, p.UpdatedAt)); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *AdministrationService) ArchiveProject(ctx context.Context, actor, workspaceID, projectID string, version int64) (*domain.Project, error) {
	wid, pid, err := twoIDs(workspaceID, projectID)
	if err != nil {
		return nil, err
	}
	if _, err = s.require(ctx, wid, actor, policy.ProjectWrite); err != nil {
		return nil, err
	}
	p, err := s.repo.GetProject(ctx, wid, pid, false)
	if err != nil {
		return nil, err
	}
	if p.Version != version || version < 1 {
		return nil, domain.ErrConflict
	}
	now := s.now().UTC()
	p.ArchivedAt, p.UpdatedAt, p.Version = &now, now, p.Version+1
	if err = s.repo.UpdateProject(ctx, p, version, audit(wid, actor, "project.archived", "project", p.ID.Hex(), nil, now)); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *AdministrationService) Invite(ctx context.Context, actor, workspaceID string, in InviteInput) (*InvitationResult, error) {
	wid, err := parseID(workspaceID, "workspaceId")
	if err != nil {
		return nil, err
	}
	member, err := s.require(ctx, wid, actor, policy.InvitationManage)
	if err != nil {
		return nil, err
	}
	email, err := validEmail(in.Email)
	if err != nil {
		return nil, err
	}
	if !in.Role.Valid() || in.Role == domain.RoleOwner {
		return nil, domain.ErrInvalidArgument
	}
	if member.Role == domain.RoleAdmin && in.Role == domain.RoleOwner {
		return nil, domain.ErrForbidden
	}
	now := s.now().UTC()
	if err = s.repo.ConsumeInvitationQuota(ctx, wid, actor, now.Truncate(time.Hour), invitationLimitPerHour); err != nil {
		return nil, err
	}
	token, tokenHash, err := invitationToken()
	if err != nil {
		return nil, err
	}
	ttl := in.ExpiresIn
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	if ttl > 30*24*time.Hour {
		return nil, domain.ErrInvalidArgument
	}
	inv := &domain.Invitation{ID: primitive.NewObjectID(), WorkspaceID: wid, Email: email, Role: in.Role, Version: 1, TokenHash: tokenHash, InvitedBy: actor, ExpiresAt: now.Add(ttl), CreatedAt: now}
	if err = s.repo.CreateInvitation(ctx, inv, audit(wid, actor, "invitation.created", "invitation", inv.ID.Hex(), map[string]any{"role": in.Role}, now)); err != nil {
		return nil, err
	}
	return &InvitationResult{Invitation: inv, Token: token}, nil
}

func (s *AdministrationService) ResendInvitation(ctx context.Context, actor, workspaceID, invitationID string) (*InvitationResult, error) {
	wid, iid, err := twoIDs(workspaceID, invitationID)
	if err != nil {
		return nil, err
	}
	if _, err = s.require(ctx, wid, actor, policy.InvitationManage); err != nil {
		return nil, err
	}
	inv, err := s.repo.GetInvitation(ctx, wid, iid)
	if err != nil {
		return nil, err
	}
	if inv.AcceptedAt != nil || inv.RevokedAt != nil {
		return nil, domain.ErrConflict
	}
	now := s.now().UTC()
	if err = s.repo.ConsumeInvitationQuota(ctx, wid, actor, now.Truncate(time.Hour), invitationLimitPerHour); err != nil {
		return nil, err
	}
	token, tokenHash, err := invitationToken()
	if err != nil {
		return nil, err
	}
	expectedVersion := inv.Version
	inv.TokenHash, inv.InvitedBy, inv.CreatedAt, inv.ExpiresAt, inv.Version = tokenHash, actor, now, now.Add(7*24*time.Hour), inv.Version+1
	if err = s.repo.UpdateInvitation(ctx, inv, expectedVersion, audit(wid, actor, "invitation.resent", "invitation", iid.Hex(), nil, now)); err != nil {
		return nil, err
	}
	return &InvitationResult{Invitation: inv, Token: token}, nil
}

func (s *AdministrationService) RevokeInvitation(ctx context.Context, actor, workspaceID, invitationID string) error {
	wid, iid, err := twoIDs(workspaceID, invitationID)
	if err != nil {
		return err
	}
	if _, err = s.require(ctx, wid, actor, policy.InvitationManage); err != nil {
		return err
	}
	inv, err := s.repo.GetInvitation(ctx, wid, iid)
	if err != nil {
		return err
	}
	if inv.AcceptedAt != nil || inv.RevokedAt != nil {
		return domain.ErrConflict
	}
	now := s.now().UTC()
	expectedVersion := inv.Version
	inv.RevokedAt, inv.RevokedBy = &now, actor
	inv.Version++
	return s.repo.UpdateInvitation(ctx, inv, expectedVersion, audit(wid, actor, "invitation.revoked", "invitation", iid.Hex(), nil, now))
}

func (s *AdministrationService) AcceptInvitation(ctx context.Context, actor, token string) (*domain.Membership, error) {
	if err := s.identities.Validate(ctx, actor); err != nil {
		return nil, err
	}
	hash := sha256.Sum256([]byte(strings.TrimSpace(token)))
	inv, err := s.repo.GetInvitationByTokenHash(ctx, hex.EncodeToString(hash[:]))
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	if inv.RevokedAt != nil || inv.AcceptedAt != nil {
		return nil, domain.ErrConflict
	}
	if !now.Before(inv.ExpiresAt) {
		return nil, domain.ErrInvitationExpired
	}
	if _, membershipErr := s.repo.GetMembership(ctx, inv.WorkspaceID, actor); membershipErr == nil {
		return nil, domain.ErrConflict
	} else if !errors.Is(membershipErr, domain.ErrNotFound) {
		return nil, membershipErr
	}
	expectedVersion := inv.Version
	inv.AcceptedAt, inv.AcceptedBy = &now, actor
	inv.Version++
	m := &domain.Membership{ID: primitive.NewObjectID(), WorkspaceID: inv.WorkspaceID, UserID: actor, Role: inv.Role, Status: domain.MembershipActive, CreatedAt: now, UpdatedAt: now}
	if err = s.repo.AcceptInvitation(ctx, inv, expectedVersion, m, audit(inv.WorkspaceID, actor, "invitation.accepted", "invitation", inv.ID.Hex(), nil, now)); err != nil {
		return nil, err
	}
	s.authorizer.Invalidate(inv.WorkspaceID, actor)
	return m, nil
}

func (s *AdministrationService) ListMembers(ctx context.Context, actor, workspaceID string) ([]*domain.Membership, error) {
	wid, err := parseID(workspaceID, "workspaceId")
	if err != nil {
		return nil, err
	}
	if _, err = s.require(ctx, wid, actor, policy.MemberRead); err != nil {
		return nil, err
	}
	return s.repo.ListMemberships(ctx, wid)
}

func (s *AdministrationService) UpdateMember(ctx context.Context, actor, workspaceID, userID string, in UpdateMemberInput) (*domain.Membership, error) {
	wid, err := parseID(workspaceID, "workspaceId")
	if err != nil {
		return nil, err
	}
	manager, err := s.require(ctx, wid, actor, policy.MemberManage)
	if err != nil {
		return nil, err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, domain.ErrInvalidArgument
	}
	target, err := s.repo.GetMembership(ctx, wid, userID)
	if err != nil {
		return nil, err
	}
	if target.Role == domain.RoleOwner || in.Role == domain.RoleOwner {
		return nil, domain.ErrForbidden
	}
	if manager.Role == domain.RoleAdmin && target.Role == domain.RoleAdmin && target.UserID != actor {
		return nil, domain.ErrForbidden
	}
	if !in.Role.Valid() || (in.Status != domain.MembershipActive && in.Status != domain.MembershipSuspended) {
		return nil, domain.ErrInvalidArgument
	}
	if target.UserID == actor && in.Status == domain.MembershipSuspended {
		return nil, domain.ErrInvalidArgument
	}
	now := s.now().UTC()
	target.Role, target.Status, target.UpdatedAt = in.Role, in.Status, now
	details := map[string]any{"role": in.Role, "status": in.Status}
	if err = s.repo.UpsertMembership(ctx, target, audit(wid, actor, "membership.updated", "membership", userID, details, now)); err != nil {
		return nil, err
	}
	s.authorizer.Invalidate(wid, userID)
	return target, nil
}

func (s *AdministrationService) RemoveMember(ctx context.Context, actor, workspaceID, userID string) error {
	wid, err := parseID(workspaceID, "workspaceId")
	if err != nil {
		return err
	}
	manager, err := s.require(ctx, wid, actor, policy.MemberManage)
	if err != nil {
		return err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" || userID == actor {
		return domain.ErrInvalidArgument
	}
	target, err := s.repo.GetMembership(ctx, wid, userID)
	if err != nil {
		return err
	}
	if target.Role == domain.RoleOwner || (manager.Role == domain.RoleAdmin && target.Role == domain.RoleAdmin) {
		return domain.ErrForbidden
	}
	now := s.now().UTC()
	// Existing task reporter/assignee IDs remain as immutable attribution; removal only revokes future access.
	details := map[string]any{"taskAttributionPreserved": true}
	if err = s.repo.RemoveMembership(ctx, wid, userID, audit(wid, actor, "membership.removed", "membership", userID, details, now)); err != nil {
		return err
	}
	s.authorizer.Invalidate(wid, userID)
	return nil
}

func (s *AdministrationService) ListAudit(ctx context.Context, actor, workspaceID string, limit int64) ([]*domain.AuditEvent, error) {
	wid, err := parseID(workspaceID, "workspaceId")
	if err != nil {
		return nil, err
	}
	if _, err = s.require(ctx, wid, actor, policy.AuditRead); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		return nil, domain.ErrInvalidArgument
	}
	return s.repo.ListAuditEvents(ctx, wid, limit)
}

func (s *AdministrationService) require(ctx context.Context, workspaceID primitive.ObjectID, actor string, action policy.Action) (*domain.Membership, error) {
	if err := s.identities.Validate(ctx, actor); err != nil {
		return nil, err
	}
	membership, err := s.authorizer.Require(ctx, workspaceID, actor, action)
	if err != nil {
		return nil, err
	}
	if action != policy.WorkspaceRead && action != policy.ProjectRead && action != policy.MemberRead && action != policy.AuditRead {
		workspace, workspaceErr := s.repo.GetWorkspace(ctx, workspaceID, true)
		if workspaceErr != nil {
			return nil, workspaceErr
		}
		if workspace.ArchivedAt != nil {
			return nil, domain.ErrConflict
		}
	}
	return membership, nil
}

func workspaceName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" || len(name) > 100 {
		return "", domain.ErrInvalidArgument
	}
	return name, nil
}
func validEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 254 {
		return "", domain.ErrInvalidArgument
	}
	return email, nil
}
func validateProject(name, key, description string, workflow []domain.WorkflowStatus, visibility domain.ProjectVisibility) error {
	if name == "" || len(name) > 100 || !projectKeyPattern.MatchString(key) || len(description) > 2000 || (visibility != domain.ProjectWorkspace && visibility != domain.ProjectPrivate) || len(workflow) == 0 || len(workflow) > 50 {
		return domain.ErrInvalidArgument
	}
	seen := map[string]bool{}
	for i := range workflow {
		workflow[i].Key = strings.TrimSpace(workflow[i].Key)
		workflow[i].Name = strings.TrimSpace(workflow[i].Name)
		state := workflow[i]
		if state.Key == "" || state.Name == "" || len(state.Key) > 50 || len(state.Name) > 100 || seen[state.Key] {
			return domain.ErrInvalidArgument
		}
		seen[state.Key] = true
	}
	return nil
}
func twoIDs(first, second string) (primitive.ObjectID, primitive.ObjectID, error) {
	a, err := parseID(first, "workspaceId")
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}
	b, err := parseID(second, "id")
	return a, b, err
}
func invitationToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(hash[:]), nil
}
func audit(workspaceID primitive.ObjectID, actor, action, resource, resourceID string, details map[string]any, now time.Time) *domain.AuditEvent {
	return &domain.AuditEvent{ID: primitive.NewObjectID(), WorkspaceID: workspaceID, ActorID: actor, Action: action, Resource: resource, ResourceID: resourceID, Details: details, CreatedAt: now}
}
