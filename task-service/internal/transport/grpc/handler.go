package grpc

import (
	"context"
	"errors"
	"time"

	taskv1 "github.com/Ishan-Gijavanekar/task-service/api/proto/task/v1"
	"github.com/Ishan-Gijavanekar/task-service/internal/auth"
	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"github.com/Ishan-Gijavanekar/task-service/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	taskv1.UnimplementedTaskServiceServer
	s *service.Service
}

func NewHandler(s *service.Service) *Handler { return &Handler{s: s} }
func actor(ctx context.Context) (string, error) {
	id, e := auth.Actor(ctx)
	if e != nil {
		return "", mapped(e)
	}
	return id, nil
}
func (h *Handler) CreateTask(ctx context.Context, r *taskv1.CreateTaskRequest) (*taskv1.TaskResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	t, e := h.s.Create(ctx, a, service.CreateInput{WorkspaceID: r.GetWorkspaceId(), ProjectID: r.GetProjectId(), Title: r.GetTitle(), Description: r.GetDescription(), Priority: domain.Priority(r.GetPriority()), AssigneeID: r.GetAssigneeId(), LabelIDs: r.GetLabelIds(), DueDate: asTime(r.GetDueDate()), StartDate: asTime(r.GetStartDate()), EstimateMinutes: r.GetEstimateMinutes(), IdempotencyKey: r.GetIdempotencyKey(), Checklist: checklistFromProto(r.GetChecklist()), Dependencies: r.GetDependencyIds()})
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.TaskResponse{Task: toProto(t)}, nil
}
func (h *Handler) GetTask(ctx context.Context, r *taskv1.GetTaskRequest) (*taskv1.TaskResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	t, e := h.s.Get(ctx, a, r.GetWorkspaceId(), r.GetId(), r.GetIncludeDeleted())
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.TaskResponse{Task: toProto(t)}, nil
}
func (h *Handler) UpdateTask(ctx context.Context, r *taskv1.UpdateTaskRequest) (*taskv1.TaskResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	t, e := h.s.Update(ctx, a, service.UpdateInput{WorkspaceID: r.GetWorkspaceId(), ID: r.GetId(), Title: r.GetTitle(), Description: r.GetDescription(), Status: domain.TaskStatus(r.GetStatus()), Priority: domain.Priority(r.GetPriority()), AssigneeID: r.GetAssigneeId(), LabelIDs: r.GetLabelIds(), DueDate: asTime(r.GetDueDate()), StartDate: asTime(r.GetStartDate()), EstimateMinutes: r.GetEstimateMinutes(), Version: r.GetVersion(), IdempotencyKey: r.GetIdempotencyKey(), Checklist: checklistFromProto(r.GetChecklist()), Dependencies: r.GetDependencyIds()})
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.TaskResponse{Task: toProto(t)}, nil
}
func (h *Handler) DeleteTask(ctx context.Context, r *taskv1.DeleteTaskRequest) (*emptypb.Empty, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	if e = h.s.Delete(ctx, a, r.GetWorkspaceId(), r.GetId(), r.GetVersion(), r.GetIdempotencyKey()); e != nil {
		return nil, mapped(e)
	}
	return &emptypb.Empty{}, nil
}
func (h *Handler) RestoreTask(ctx context.Context, r *taskv1.RestoreTaskRequest) (*taskv1.TaskResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	t, e := h.s.Restore(ctx, a, r.GetWorkspaceId(), r.GetId(), r.GetVersion(), r.GetIdempotencyKey())
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.TaskResponse{Task: toProto(t)}, nil
}
func (h *Handler) ReorderTask(ctx context.Context, r *taskv1.ReorderTaskRequest) (*taskv1.TaskResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	var b, after *float64
	if r.BeforePosition != nil {
		v := r.GetBeforePosition()
		b = &v
	}
	if r.AfterPosition != nil {
		v := r.GetAfterPosition()
		after = &v
	}
	t, e := h.s.Reorder(ctx, a, r.GetWorkspaceId(), r.GetId(), r.GetVersion(), b, after, r.GetIdempotencyKey())
	if e != nil {
		return nil, mapped(e)
	}
	return &taskv1.TaskResponse{Task: toProto(t)}, nil
}
func (h *Handler) ListTasks(ctx context.Context, r *taskv1.ListTasksRequest) (*taskv1.ListTasksResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	in := service.ListInput{WorkspaceID: r.GetWorkspaceId(), ProjectID: r.GetProjectId(), AssigneeID: r.GetAssigneeId(), ReporterID: r.GetReporterId(), LabelID: r.GetLabelId(), Query: r.GetQuery(), Limit: int64(r.GetLimit()), Cursor: r.GetCursor(), IncludeDeleted: r.GetIncludeDeleted(), DueFrom: asTime(r.GetDueFrom()), DueTo: asTime(r.GetDueTo())}
	for _, v := range r.GetStatuses() {
		in.Statuses = append(in.Statuses, domain.TaskStatus(v))
	}
	for _, v := range r.GetPriorities() {
		in.Priorities = append(in.Priorities, domain.Priority(v))
	}
	result, e := h.s.List(ctx, a, in)
	if e != nil {
		return nil, mapped(e)
	}
	out := &taskv1.ListTasksResponse{NextCursor: result.NextCursor}
	for _, t := range result.Tasks {
		out.Tasks = append(out.Tasks, toProto(t))
	}
	return out, nil
}
func (h *Handler) BatchUpdateTasks(ctx context.Context, r *taskv1.BatchUpdateTasksRequest) (*taskv1.BatchUpdateTasksResponse, error) {
	a, e := actor(ctx)
	if e != nil {
		return nil, e
	}
	items := make([]service.BatchItem, 0, len(r.GetItems()))
	for _, i := range r.GetItems() {
		items = append(items, service.BatchItem{ID: i.GetId(), Status: domain.TaskStatus(i.GetStatus()), Priority: domain.Priority(i.GetPriority()), AssigneeID: i.GetAssigneeId(), Version: i.GetVersion()})
	}
	tasks, e := h.s.BatchUpdate(ctx, a, r.GetWorkspaceId(), r.GetIdempotencyKey(), items)
	if e != nil {
		return nil, mapped(e)
	}
	out := &taskv1.BatchUpdateTasksResponse{}
	for _, t := range tasks {
		out.Tasks = append(out.Tasks, toProto(t))
	}
	return out, nil
}
func asTime(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	v := t.AsTime()
	return &v
}
func stamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
func toProto(t *domain.Task) *taskv1.Task {
	labels := make([]string, 0, len(t.LabelIDs))
	for _, id := range t.LabelIDs {
		labels = append(labels, id.Hex())
	}
	out := &taskv1.Task{Id: t.ID.Hex(), WorkspaceId: t.WorkspaceID.Hex(), ProjectId: t.ProjectID.Hex(), Title: t.Title, Description: t.Description, Status: string(t.Status), Priority: string(t.Priority), AssigneeId: t.AssigneeID, ReporterId: t.ReporterID, LabelIds: labels, DueDate: stamp(t.DueDate), StartDate: stamp(t.StartDate), EstimateMinutes: t.EstimateMinutes, Position: t.Position, Version: t.Version, CreatedAt: timestamppb.New(t.CreatedAt), UpdatedAt: timestamppb.New(t.UpdatedAt), DeletedAt: stamp(t.DeletedAt)}
	for _, item := range t.Checklist {
		out.Checklist = append(out.Checklist, &taskv1.ChecklistItem{Id: item.ID.Hex(), Text: item.Text, Completed: item.Completed, Position: item.Position})
	}
	for _, dep := range t.Dependencies {
		out.Dependencies = append(out.Dependencies, &taskv1.TaskDependency{TaskId: dep.TaskID.Hex(), DependsOnTaskId: dep.DependsOnTaskID.Hex()})
	}
	return out
}

func checklistFromProto(items []*taskv1.ChecklistItem) []domain.ChecklistItem {
	out := make([]domain.ChecklistItem, 0, len(items))
	for _, item := range items {
		id := primitive.NewObjectID()
		if parsed, e := primitive.ObjectIDFromHex(item.GetId()); e == nil {
			id = parsed
		}
		out = append(out, domain.ChecklistItem{ID: id, Text: item.GetText(), Completed: item.GetCompleted(), Position: item.GetPosition()})
	}
	return out
}
func mapped(e error) error {
	code := codes.Internal
	switch {
	case errors.Is(e, domain.ErrInvalidArgument), errors.Is(e, domain.ErrInvalidTransition), errors.Is(e, domain.ErrDependencyCycle):
		code = codes.InvalidArgument
	case errors.Is(e, domain.ErrUnauthenticated):
		code = codes.Unauthenticated
	case errors.Is(e, domain.ErrForbidden):
		code = codes.PermissionDenied
	case errors.Is(e, domain.ErrNotFound):
		code = codes.NotFound
	case errors.Is(e, domain.ErrConflict), errors.Is(e, domain.ErrIdempotencyConflict):
		code = codes.Aborted
	case errors.Is(e, domain.ErrInvitationExpired):
		code = codes.FailedPrecondition
	case errors.Is(e, domain.ErrRateLimited):
		code = codes.ResourceExhausted
	}
	message := e.Error()
	if code == codes.Internal {
		message = "internal server error"
	}
	return status.Error(code, message)
}
