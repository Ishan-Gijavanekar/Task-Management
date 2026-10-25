package grpc

import (
	"context"

	taskv1 "github.com/Ishan-Gijavanekar/task-service/api/proto/task/v1"
	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"github.com/Ishan-Gijavanekar/task-service/internal/service"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AdministrationHandler struct {
	taskv1.UnimplementedWorkspaceServiceServer
	s *service.AdministrationService
}

func NewAdministrationHandler(s *service.AdministrationService) *AdministrationHandler {
	return &AdministrationHandler{s: s}
}

func (h *AdministrationHandler) CreateWorkspace(ctx context.Context, r *taskv1.CreateWorkspaceRequest) (*taskv1.WorkspaceResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	out, e := h.s.CreateWorkspace(ctx, a, service.CreateWorkspaceInput{Name: r.GetName()})
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.WorkspaceResponse{Workspace: workspaceProto(out)}, nil
}
func (h *AdministrationHandler) GetWorkspace(ctx context.Context, r *taskv1.GetWorkspaceRequest) (*taskv1.WorkspaceResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	out, e := h.s.GetWorkspace(ctx, a, r.GetWorkspaceId(), r.GetIncludeArchived())
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.WorkspaceResponse{Workspace: workspaceProto(out)}, nil
}
func (h *AdministrationHandler) UpdateWorkspace(ctx context.Context, r *taskv1.UpdateWorkspaceRequest) (*taskv1.WorkspaceResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	out, e := h.s.UpdateWorkspace(ctx, a, r.GetWorkspaceId(), service.UpdateWorkspaceInput{Name: r.GetName(), Version: r.GetVersion()})
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.WorkspaceResponse{Workspace: workspaceProto(out)}, nil
}
func (h *AdministrationHandler) ArchiveWorkspace(ctx context.Context, r *taskv1.ArchiveWorkspaceRequest) (*taskv1.WorkspaceResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	out, e := h.s.ArchiveWorkspace(ctx, a, r.GetWorkspaceId(), r.GetVersion())
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.WorkspaceResponse{Workspace: workspaceProto(out)}, nil
}
func (h *AdministrationHandler) TransferOwnership(ctx context.Context, r *taskv1.TransferOwnershipRequest) (*taskv1.WorkspaceResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	out, e := h.s.TransferOwnership(ctx, a, r.GetWorkspaceId(), r.GetNewOwnerId(), r.GetVersion())
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.WorkspaceResponse{Workspace: workspaceProto(out)}, nil
}
func (h *AdministrationHandler) CreateProject(ctx context.Context, r *taskv1.CreateProjectRequest) (*taskv1.ProjectResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	out, e := h.s.CreateProject(ctx, a, r.GetWorkspaceId(), service.CreateProjectInput{Name: r.GetName(), Key: r.GetKey(), Description: r.GetDescription(), Workflow: workflowDomain(r.GetWorkflow()), Visibility: domain.ProjectVisibility(r.GetVisibility())})
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.ProjectResponse{Project: projectProto(out)}, nil
}
func (h *AdministrationHandler) GetProject(ctx context.Context, r *taskv1.GetProjectRequest) (*taskv1.ProjectResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	out, e := h.s.GetProject(ctx, a, r.GetWorkspaceId(), r.GetProjectId(), r.GetIncludeArchived())
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.ProjectResponse{Project: projectProto(out)}, nil
}
func (h *AdministrationHandler) ListProjects(ctx context.Context, r *taskv1.ListProjectsRequest) (*taskv1.ListProjectsResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	items, e := h.s.ListProjects(ctx, a, r.GetWorkspaceId(), r.GetIncludeArchived())
	if e != nil {
		return nil, mapped(e)
	}
	out := &taskv1.ListProjectsResponse{}
	for _, item := range items {
		out.Projects = append(out.Projects, projectProto(item))
	}
	return out, nil
}
func (h *AdministrationHandler) UpdateProject(ctx context.Context, r *taskv1.UpdateProjectRequest) (*taskv1.ProjectResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	out, e := h.s.UpdateProject(ctx, a, r.GetWorkspaceId(), r.GetProjectId(), service.UpdateProjectInput{Name: r.GetName(), Description: r.GetDescription(), Workflow: workflowDomain(r.GetWorkflow()), Visibility: domain.ProjectVisibility(r.GetVisibility()), Version: r.GetVersion()})
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.ProjectResponse{Project: projectProto(out)}, nil
}
func (h *AdministrationHandler) ArchiveProject(ctx context.Context, r *taskv1.ArchiveProjectRequest) (*taskv1.ProjectResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	out, e := h.s.ArchiveProject(ctx, a, r.GetWorkspaceId(), r.GetProjectId(), r.GetVersion())
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.ProjectResponse{Project: projectProto(out)}, nil
}
func (h *AdministrationHandler) InviteMember(ctx context.Context, r *taskv1.InviteMemberRequest) (*taskv1.InvitationResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	out, e := h.s.Invite(ctx, a, r.GetWorkspaceId(), service.InviteInput{Email: r.GetEmail(), Role: domain.WorkspaceRole(r.GetRole())})
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.InvitationResponse{Invitation: invitationProto(out.Invitation), Token: out.Token}, nil
}
func (h *AdministrationHandler) ResendInvitation(ctx context.Context, r *taskv1.InvitationRequest) (*taskv1.InvitationResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	out, e := h.s.ResendInvitation(ctx, a, r.GetWorkspaceId(), r.GetInvitationId())
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.InvitationResponse{Invitation: invitationProto(out.Invitation), Token: out.Token}, nil
}
func (h *AdministrationHandler) RevokeInvitation(ctx context.Context, r *taskv1.InvitationRequest) (*emptypb.Empty, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	if e = h.s.RevokeInvitation(ctx, a, r.GetWorkspaceId(), r.GetInvitationId()); e != nil {
		return nil, mapped(e)
	}
	return &emptypb.Empty{}, nil
}
func (h *AdministrationHandler) AcceptInvitation(ctx context.Context, r *taskv1.AcceptInvitationRequest) (*taskv1.MembershipResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	out, e := h.s.AcceptInvitation(ctx, a, r.GetToken())
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.MembershipResponse{Membership: membershipProto(out)}, nil
}
func (h *AdministrationHandler) ListMembers(ctx context.Context, r *taskv1.ListMembersRequest) (*taskv1.ListMembersResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	items, e := h.s.ListMembers(ctx, a, r.GetWorkspaceId())
	if e != nil {
		return nil, mapped(e)
	}
	out := &taskv1.ListMembersResponse{}
	for _, item := range items {
		out.Members = append(out.Members, membershipProto(item))
	}
	return out, nil
}
func (h *AdministrationHandler) UpdateMember(ctx context.Context, r *taskv1.UpdateMemberRequest) (*taskv1.MembershipResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	out, e := h.s.UpdateMember(ctx, a, r.GetWorkspaceId(), r.GetUserId(), service.UpdateMemberInput{Role: domain.WorkspaceRole(r.GetRole()), Status: domain.MembershipStatus(r.GetStatus())})
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.MembershipResponse{Membership: membershipProto(out)}, nil
}
func (h *AdministrationHandler) RemoveMember(ctx context.Context, r *taskv1.RemoveMemberRequest) (*emptypb.Empty, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	if e = h.s.RemoveMember(ctx, a, r.GetWorkspaceId(), r.GetUserId()); e != nil {
		return nil, mapped(e)
	}
	return &emptypb.Empty{}, nil
}
func (h *AdministrationHandler) ListAuditEvents(ctx context.Context, r *taskv1.ListAuditEventsRequest) (*taskv1.ListAuditEventsResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	items, e := h.s.ListAudit(ctx, a, r.GetWorkspaceId(), int64(r.GetLimit()))
	if e != nil {
		return nil, mapped(e)
	}
	out := &taskv1.ListAuditEventsResponse{}
	for _, item := range items {
		details, _ := structpb.NewStruct(item.Details)
		out.Events = append(out.Events, &taskv1.AuditEvent{Id: item.ID.Hex(), WorkspaceId: item.WorkspaceID.Hex(), ActorId: item.ActorID, Action: item.Action, Resource: item.Resource, ResourceId: item.ResourceID, Details: details, CreatedAt: timestamppb.New(item.CreatedAt)})
	}
	return out, nil
}

func workspaceProto(w *domain.Workspace) *taskv1.Workspace {
	return &taskv1.Workspace{Id: w.ID.Hex(), Name: w.Name, OwnerId: w.OwnerID, Version: w.Version, CreatedAt: timestamppb.New(w.CreatedAt), UpdatedAt: timestamppb.New(w.UpdatedAt), ArchivedAt: stamp(w.ArchivedAt)}
}
func projectProto(p *domain.Project) *taskv1.Project {
	out := &taskv1.Project{Id: p.ID.Hex(), WorkspaceId: p.WorkspaceID.Hex(), Name: p.Name, Key: p.Key, Description: p.Description, Visibility: string(p.Visibility), Version: p.Version, CreatedAt: timestamppb.New(p.CreatedAt), UpdatedAt: timestamppb.New(p.UpdatedAt), ArchivedAt: stamp(p.ArchivedAt)}
	for _, state := range p.Workflow {
		out.Workflow = append(out.Workflow, &taskv1.WorkflowStatus{Key: state.Key, Name: state.Name})
	}
	return out
}
func workflowDomain(items []*taskv1.WorkflowStatus) []domain.WorkflowStatus {
	out := make([]domain.WorkflowStatus, 0, len(items))
	for _, item := range items {
		out = append(out, domain.WorkflowStatus{Key: item.GetKey(), Name: item.GetName()})
	}
	return out
}
func membershipProto(m *domain.Membership) *taskv1.Membership {
	return &taskv1.Membership{Id: m.ID.Hex(), WorkspaceId: m.WorkspaceID.Hex(), UserId: m.UserID, Role: string(m.Role), Status: string(m.Status), CreatedAt: timestamppb.New(m.CreatedAt), UpdatedAt: timestamppb.New(m.UpdatedAt)}
}
func invitationProto(i *domain.Invitation) *taskv1.Invitation {
	return &taskv1.Invitation{Id: i.ID.Hex(), WorkspaceId: i.WorkspaceID.Hex(), Email: i.Email, Role: string(i.Role), InvitedBy: i.InvitedBy, ExpiresAt: timestamppb.New(i.ExpiresAt), CreatedAt: timestamppb.New(i.CreatedAt), AcceptedAt: stamp(i.AcceptedAt), AcceptedBy: i.AcceptedBy, RevokedAt: stamp(i.RevokedAt), RevokedBy: i.RevokedBy, Version: i.Version}
}
