package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type memoryAdministration struct {
	workspaces  map[primitive.ObjectID]*domain.Workspace
	projects    map[primitive.ObjectID]*domain.Project
	members     map[string]*domain.Membership
	invitations map[primitive.ObjectID]*domain.Invitation
	quotas      map[string]int64
	audits      []*domain.AuditEvent
}

func newMemoryAdministration() *memoryAdministration {
	return &memoryAdministration{workspaces: map[primitive.ObjectID]*domain.Workspace{}, projects: map[primitive.ObjectID]*domain.Project{}, members: map[string]*domain.Membership{}, invitations: map[primitive.ObjectID]*domain.Invitation{}, quotas: map[string]int64{}}
}
func memberKey(w primitive.ObjectID, user string) string { return w.Hex() + ":" + user }
func (m *memoryAdministration) CreateWorkspace(_ context.Context, w *domain.Workspace, member *domain.Membership, audit *domain.AuditEvent) error {
	m.workspaces[w.ID] = w
	m.members[memberKey(w.ID, member.UserID)] = member
	m.audits = append(m.audits, audit)
	return nil
}
func (m *memoryAdministration) GetWorkspace(_ context.Context, id primitive.ObjectID, archived bool) (*domain.Workspace, error) {
	w := m.workspaces[id]
	if w == nil || (!archived && w.ArchivedAt != nil) {
		return nil, domain.ErrNotFound
	}
	c := *w
	return &c, nil
}
func (m *memoryAdministration) UpdateWorkspace(_ context.Context, w *domain.Workspace, v int64, a *domain.AuditEvent) error {
	old := m.workspaces[w.ID]
	if old == nil {
		return domain.ErrNotFound
	}
	if old.Version != v {
		return domain.ErrConflict
	}
	c := *w
	m.workspaces[w.ID] = &c
	m.audits = append(m.audits, a)
	return nil
}
func (m *memoryAdministration) TransferOwnership(_ context.Context, w *domain.Workspace, v int64, old, next *domain.Membership, a *domain.AuditEvent) error {
	if m.workspaces[w.ID].Version != v {
		return domain.ErrConflict
	}
	m.workspaces[w.ID] = w
	m.members[memberKey(w.ID, old.UserID)] = old
	m.members[memberKey(w.ID, next.UserID)] = next
	m.audits = append(m.audits, a)
	return nil
}
func (m *memoryAdministration) CreateProject(_ context.Context, p *domain.Project, a *domain.AuditEvent) error {
	for _, x := range m.projects {
		if x.WorkspaceID == p.WorkspaceID && x.Key == p.Key {
			return domain.ErrConflict
		}
	}
	m.projects[p.ID] = p
	m.audits = append(m.audits, a)
	return nil
}
func (m *memoryAdministration) GetProject(_ context.Context, w, id primitive.ObjectID, archived bool) (*domain.Project, error) {
	p := m.projects[id]
	if p == nil || p.WorkspaceID != w || (!archived && p.ArchivedAt != nil) {
		return nil, domain.ErrNotFound
	}
	c := *p
	return &c, nil
}
func (m *memoryAdministration) ListProjects(_ context.Context, w primitive.ObjectID, archived bool) ([]*domain.Project, error) {
	var out []*domain.Project
	for _, p := range m.projects {
		if p.WorkspaceID == w && (archived || p.ArchivedAt == nil) {
			c := *p
			out = append(out, &c)
		}
	}
	return out, nil
}
func (m *memoryAdministration) UpdateProject(_ context.Context, p *domain.Project, v int64, a *domain.AuditEvent) error {
	old := m.projects[p.ID]
	if old == nil {
		return domain.ErrNotFound
	}
	if old.Version != v {
		return domain.ErrConflict
	}
	m.projects[p.ID] = p
	m.audits = append(m.audits, a)
	return nil
}
func (m *memoryAdministration) GetMembership(_ context.Context, w primitive.ObjectID, u string) (*domain.Membership, error) {
	x := m.members[memberKey(w, u)]
	if x == nil {
		return nil, domain.ErrNotFound
	}
	c := *x
	return &c, nil
}
func (m *memoryAdministration) ListMemberships(_ context.Context, w primitive.ObjectID) ([]*domain.Membership, error) {
	var out []*domain.Membership
	for _, x := range m.members {
		if x.WorkspaceID == w {
			c := *x
			out = append(out, &c)
		}
	}
	return out, nil
}
func (m *memoryAdministration) UpsertMembership(_ context.Context, x *domain.Membership, a *domain.AuditEvent) error {
	c := *x
	m.members[memberKey(x.WorkspaceID, x.UserID)] = &c
	m.audits = append(m.audits, a)
	return nil
}
func (m *memoryAdministration) RemoveMembership(_ context.Context, w primitive.ObjectID, u string, a *domain.AuditEvent) error {
	k := memberKey(w, u)
	if m.members[k] == nil {
		return domain.ErrNotFound
	}
	delete(m.members, k)
	m.audits = append(m.audits, a)
	return nil
}
func (m *memoryAdministration) CreateInvitation(_ context.Context, i *domain.Invitation, a *domain.AuditEvent) error {
	m.invitations[i.ID] = i
	m.audits = append(m.audits, a)
	return nil
}
func (m *memoryAdministration) GetInvitation(_ context.Context, w, id primitive.ObjectID) (*domain.Invitation, error) {
	i := m.invitations[id]
	if i == nil || i.WorkspaceID != w {
		return nil, domain.ErrNotFound
	}
	c := *i
	return &c, nil
}
func (m *memoryAdministration) GetInvitationByTokenHash(_ context.Context, h string) (*domain.Invitation, error) {
	for _, i := range m.invitations {
		if i.TokenHash == h {
			c := *i
			return &c, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (m *memoryAdministration) UpdateInvitation(_ context.Context, i *domain.Invitation, version int64, a *domain.AuditEvent) error {
	old := m.invitations[i.ID]
	if old == nil || old.Version != version || old.AcceptedAt != nil || old.RevokedAt != nil {
		return domain.ErrConflict
	}
	c := *i
	m.invitations[i.ID] = &c
	m.audits = append(m.audits, a)
	return nil
}
func (m *memoryAdministration) AcceptInvitation(_ context.Context, i *domain.Invitation, version int64, x *domain.Membership, a *domain.AuditEvent) error {
	old := m.invitations[i.ID]
	if old == nil || old.Version != version || old.AcceptedAt != nil || old.RevokedAt != nil {
		return domain.ErrConflict
	}
	m.invitations[i.ID] = i
	m.members[memberKey(x.WorkspaceID, x.UserID)] = x
	m.audits = append(m.audits, a)
	return nil
}
func (m *memoryAdministration) ConsumeInvitationQuota(_ context.Context, w primitive.ObjectID, a string, bucket time.Time, limit int64) error {
	key := w.Hex() + a + bucket.Format(time.RFC3339)
	m.quotas[key]++
	if m.quotas[key] > limit {
		return domain.ErrRateLimited
	}
	return nil
}
func (m *memoryAdministration) ListAuditEvents(_ context.Context, w primitive.ObjectID, limit int64) ([]*domain.AuditEvent, error) {
	var out []*domain.AuditEvent
	for i := len(m.audits) - 1; i >= 0 && int64(len(out)) < limit; i-- {
		if m.audits[i].WorkspaceID == w {
			out = append(out, m.audits[i])
		}
	}
	return out, nil
}

func setupAdministration(t *testing.T) (*AdministrationService, *memoryAdministration, *Authorizer) {
	t.Helper()
	repo := newMemoryAdministration()
	authz := NewAuthorizer(repo, time.Minute)
	svc, err := NewAdministration(repo, StaticIdentityValidator{}, authz)
	if err != nil {
		t.Fatal(err)
	}
	fixed := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return fixed }
	authz.now = svc.now
	return svc, repo, authz
}

func TestWorkspaceLifecycleAndOwnershipTransfer(t *testing.T) {
	s, repo, _ := setupAdministration(t)
	ctx := context.Background()
	w, err := s.CreateWorkspace(ctx, "owner", CreateWorkspaceInput{Name: " Team "})
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "Team" || w.Version != 1 {
		t.Fatalf("unexpected workspace: %+v", w)
	}
	repo.members[memberKey(w.ID, "next")] = &domain.Membership{ID: primitive.NewObjectID(), WorkspaceID: w.ID, UserID: "next", Role: domain.RoleMember, Status: domain.MembershipActive, CreatedAt: s.now(), UpdatedAt: s.now()}
	w, err = s.UpdateWorkspace(ctx, "owner", w.ID.Hex(), UpdateWorkspaceInput{Name: "Renamed", Version: 1})
	if err != nil {
		t.Fatal(err)
	}
	w, err = s.TransferOwnership(ctx, "owner", w.ID.Hex(), "next", w.Version)
	if err != nil {
		t.Fatal(err)
	}
	if w.OwnerID != "next" || repo.members[memberKey(w.ID, "owner")].Role != domain.RoleAdmin || repo.members[memberKey(w.ID, "next")].Role != domain.RoleOwner {
		t.Fatal("ownership roles were not transferred atomically")
	}
	if _, err = s.ArchiveWorkspace(ctx, "owner", w.ID.Hex(), w.Version); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("old owner should no longer archive: %v", err)
	}
	if _, err = s.ArchiveWorkspace(ctx, "next", w.ID.Hex(), w.Version); err != nil {
		t.Fatal(err)
	}
}

func TestProjectValidationAuthorizationAndTenantIsolation(t *testing.T) {
	s, repo, _ := setupAdministration(t)
	ctx := context.Background()
	w, _ := s.CreateWorkspace(ctx, "owner", CreateWorkspaceInput{Name: "Team"})
	repo.members[memberKey(w.ID, "viewer")] = &domain.Membership{ID: primitive.NewObjectID(), WorkspaceID: w.ID, UserID: "viewer", Role: domain.RoleViewer, Status: domain.MembershipActive}
	if _, err := s.CreateProject(ctx, "viewer", w.ID.Hex(), CreateProjectInput{Name: "Roadmap", Key: "RD"}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("viewer created project: %v", err)
	}
	p, err := s.CreateProject(ctx, "owner", w.ID.Hex(), CreateProjectInput{Name: "Roadmap", Key: "rd"})
	if err != nil {
		t.Fatal(err)
	}
	if p.Key != "RD" || len(p.Workflow) == 0 {
		t.Fatal("project defaults missing")
	}
	other, _ := s.CreateWorkspace(ctx, "other", CreateWorkspaceInput{Name: "Other"})
	if _, err = s.GetProject(ctx, "other", other.ID.Hex(), p.ID.Hex(), false); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("cross-workspace project lookup leaked: %v", err)
	}
}

func TestInvitationAcceptIsSingleUseAndInvalidatesMembershipCache(t *testing.T) {
	s, _, authz := setupAdministration(t)
	ctx := context.Background()
	w, _ := s.CreateWorkspace(ctx, "owner", CreateWorkspaceInput{Name: "Team"})
	result, err := s.Invite(ctx, "owner", w.ID.Hex(), InviteInput{Email: "new@example.com", Role: domain.RoleMember})
	if err != nil {
		t.Fatal(err)
	}
	member, err := s.AcceptInvitation(ctx, "new-user", result.Token)
	if err != nil {
		t.Fatal(err)
	}
	if member.Role != domain.RoleMember {
		t.Fatal("wrong accepted role")
	}
	if _, err = s.AcceptInvitation(ctx, "second-user", result.Token); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("token was reusable: %v", err)
	}
	if _, err = authz.Membership(ctx, w.ID, "new-user"); err != nil {
		t.Fatalf("accepted membership unavailable: %v", err)
	}
}

func TestInvitationCannotReplaceExistingMembership(t *testing.T) {
	s, repo, _ := setupAdministration(t)
	ctx := context.Background()
	w, _ := s.CreateWorkspace(ctx, "owner", CreateWorkspaceInput{Name: "Team"})
	result, err := s.Invite(ctx, "owner", w.ID.Hex(), InviteInput{Email: "owner@example.com", Role: domain.RoleViewer})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.AcceptInvitation(ctx, "owner", result.Token); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("existing owner membership was replaceable: %v", err)
	}
	if repo.members[memberKey(w.ID, "owner")].Role != domain.RoleOwner {
		t.Fatal("owner role changed")
	}
}

func TestInvitationExpiryRevocationAndRateLimit(t *testing.T) {
	s, repo, _ := setupAdministration(t)
	ctx := context.Background()
	w, _ := s.CreateWorkspace(ctx, "owner", CreateWorkspaceInput{Name: "Team"})
	expired, err := s.Invite(ctx, "owner", w.ID.Hex(), InviteInput{Email: "expired@example.com", Role: domain.RoleViewer, ExpiresIn: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	s.now = func() time.Time { return time.Date(2026, 10, 25, 14, 0, 0, 0, time.UTC) }
	if _, err = s.AcceptInvitation(ctx, "late", expired.Token); !errors.Is(err, domain.ErrInvitationExpired) {
		t.Fatalf("got %v", err)
	}
	s.now = func() time.Time { return time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC) }
	revoked, err := s.Invite(ctx, "owner", w.ID.Hex(), InviteInput{Email: "revoked@example.com", Role: domain.RoleViewer})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.RevokeInvitation(ctx, "owner", w.ID.Hex(), revoked.Invitation.ID.Hex()); err != nil {
		t.Fatal(err)
	}
	if _, err = s.AcceptInvitation(ctx, "revoked", revoked.Token); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("revoked invite accepted: %v", err)
	}
	for i := len(repo.invitations); i < int(invitationLimitPerHour); i++ {
		_, err = s.Invite(ctx, "owner", w.ID.Hex(), InviteInput{Email: time.Unix(int64(i), 0).Format("150405") + "@example.com", Role: domain.RoleViewer})
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = s.Invite(ctx, "owner", w.ID.Hex(), InviteInput{Email: "limit@example.com", Role: domain.RoleViewer}); !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("got %v", err)
	}
}

func TestMemberProtectionSuspensionAndRemovalAudit(t *testing.T) {
	s, repo, authz := setupAdministration(t)
	ctx := context.Background()
	w, _ := s.CreateWorkspace(ctx, "owner", CreateWorkspaceInput{Name: "Team"})
	now := s.now()
	repo.members[memberKey(w.ID, "admin")] = &domain.Membership{ID: primitive.NewObjectID(), WorkspaceID: w.ID, UserID: "admin", Role: domain.RoleAdmin, Status: domain.MembershipActive, CreatedAt: now, UpdatedAt: now}
	repo.members[memberKey(w.ID, "member")] = &domain.Membership{ID: primitive.NewObjectID(), WorkspaceID: w.ID, UserID: "member", Role: domain.RoleMember, Status: domain.MembershipActive, CreatedAt: now, UpdatedAt: now}
	if _, err := s.UpdateMember(ctx, "admin", w.ID.Hex(), "owner", UpdateMemberInput{Role: domain.RoleMember, Status: domain.MembershipActive}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("admin changed owner: %v", err)
	}
	if _, err := s.UpdateMember(ctx, "owner", w.ID.Hex(), "member", UpdateMemberInput{Role: domain.RoleViewer, Status: domain.MembershipSuspended}); err != nil {
		t.Fatal(err)
	}
	if _, err := authz.Membership(ctx, w.ID, "member"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ListMembers(ctx, "member", w.ID.Hex()); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("suspended member retained access: %v", err)
	}
	if err := s.RemoveMember(ctx, "owner", w.ID.Hex(), "admin"); err != nil {
		t.Fatal(err)
	}
	if repo.members[memberKey(w.ID, "admin")] != nil {
		t.Fatal("member not removed")
	}
	events, err := s.ListAudit(ctx, "owner", w.ID.Hex(), 100)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, event := range events {
		if event.Action == "membership.removed" && event.Details["taskAttributionPreserved"] == true {
			found = true
		}
	}
	if !found {
		t.Fatal("removal audit implication missing")
	}
}
