package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"github.com/Ishan-Gijavanekar/task-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IdentityValidator interface {
	Validate(context.Context, string) error
}
type StaticIdentityValidator struct{}

func (StaticIdentityValidator) Validate(_ context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return domain.ErrUnauthenticated
	}
	return nil
}

type Service struct {
	repo       repository.TaskRepository
	identities IdentityValidator
	now        func() time.Time
}

func New(repo repository.TaskRepository, identities IdentityValidator) (*Service, error) {
	if repo == nil || identities == nil {
		return nil, fmt.Errorf("repository and identity validator are required")
	}
	return &Service{repo: repo, identities: identities, now: time.Now}, nil
}

type CreateInput struct {
	WorkspaceID, ProjectID, Title, Description, AssigneeID string
	Priority                                               domain.Priority
	LabelIDs                                               []string
	DueDate, StartDate                                     *time.Time
	EstimateMinutes                                        int64
	Position                                               *float64
	Checklist                                              []domain.ChecklistItem
	Dependencies                                           []string
	IdempotencyKey                                         string
}
type UpdateInput struct {
	WorkspaceID, ID, Title, Description, AssigneeID string
	Status                                          domain.TaskStatus
	Priority                                        domain.Priority
	LabelIDs                                        []string
	DueDate, StartDate                              *time.Time
	EstimateMinutes                                 int64
	Checklist                                       []domain.ChecklistItem
	Dependencies                                    []string
	Version                                         int64
	IdempotencyKey                                  string
}
type ListInput struct {
	WorkspaceID, ProjectID, AssigneeID, ReporterID, LabelID, Query, Cursor string
	Statuses                                                               []domain.TaskStatus
	Priorities                                                             []domain.Priority
	DueFrom, DueTo                                                         *time.Time
	IncludeDeleted                                                         bool
	Limit                                                                  int64
}
type ListResult struct {
	Tasks      []*domain.Task `json:"tasks"`
	NextCursor string         `json:"nextCursor,omitempty"`
}
type BatchItem struct {
	ID, AssigneeID string
	Status         domain.TaskStatus
	Priority       domain.Priority
	Version        int64
}

func (s *Service) Create(ctx context.Context, actor string, in CreateInput) (*domain.Task, error) {
	wid, e := parseID(in.WorkspaceID, "workspaceId")
	if e != nil {
		return nil, e
	}
	pid, e := parseID(in.ProjectID, "projectId")
	if e != nil {
		return nil, e
	}
	if e = s.authorize(ctx, wid, actor); e != nil {
		return nil, e
	}
	title := strings.TrimSpace(in.Title)
	if title == "" || len(title) > 300 {
		return nil, fmt.Errorf("%w: title", domain.ErrInvalidArgument)
	}
	p := in.Priority
	if p == "" {
		p = domain.PriorityNone
	}
	if !p.Valid() {
		return nil, domain.ErrInvalidArgument
	}
	labels, e := parseIDs(in.LabelIDs, "labelIds")
	if e != nil {
		return nil, e
	}
	deps, e := s.dependencies(ctx, wid, primitive.NilObjectID, in.Dependencies)
	if e != nil {
		return nil, e
	}
	now := s.now().UTC()
	pos := domain.PositionBetween(nil, nil)
	if in.Position != nil {
		pos = *in.Position
	}
	t := &domain.Task{ID: primitive.NewObjectID(), WorkspaceID: wid, ProjectID: pid, Title: title, Description: strings.TrimSpace(in.Description), Status: domain.StatusBacklog, Priority: p, AssigneeID: strings.TrimSpace(in.AssigneeID), ReporterID: actor, LabelIDs: labels, Checklist: in.Checklist, Dependencies: deps, DueDate: in.DueDate, StartDate: in.StartDate, EstimateMinutes: in.EstimateMinutes, Position: pos, Version: 1, CreatedAt: now, UpdatedAt: now}
	if e = validate(t); e != nil {
		return nil, e
	}
	h := hash(in)
	if replay, e := s.replay(ctx, wid, actor, in.IdempotencyKey, h); e != nil || replay != nil {
		return replay, e
	}
	return s.repo.Apply(ctx, repository.Mutation{Task: t, Create: true, Event: event(t, actor, "task.created", now), Idempotency: idem(wid, actor, in.IdempotencyKey, h, now)})
}
func (s *Service) Get(ctx context.Context, actor, workspace, id string, deleted bool) (*domain.Task, error) {
	wid, e := parseID(workspace, "workspaceId")
	if e != nil {
		return nil, e
	}
	tid, e := parseID(id, "id")
	if e != nil {
		return nil, e
	}
	if e = s.authorize(ctx, wid, actor); e != nil {
		return nil, e
	}
	return s.repo.Get(ctx, wid, tid, deleted)
}
func (s *Service) Update(ctx context.Context, actor string, in UpdateInput) (*domain.Task, error) {
	wid, e := parseID(in.WorkspaceID, "workspaceId")
	if e != nil {
		return nil, e
	}
	tid, e := parseID(in.ID, "id")
	if e != nil {
		return nil, e
	}
	if e = s.authorize(ctx, wid, actor); e != nil {
		return nil, e
	}
	h := hash(in)
	if replay, e := s.replay(ctx, wid, actor, in.IdempotencyKey, h); e != nil || replay != nil {
		return replay, e
	}
	old, e := s.repo.Get(ctx, wid, tid, false)
	if e != nil {
		return nil, e
	}
	if old.Version != in.Version {
		return nil, domain.ErrConflict
	}
	if in.Status == "" {
		in.Status = old.Status
	}
	if !in.Status.Valid() {
		return nil, domain.ErrInvalidArgument
	}
	if !domain.CanTransition(old.Status, in.Status) {
		return nil, domain.ErrInvalidTransition
	}
	if in.Priority == "" {
		in.Priority = old.Priority
	}
	if !in.Priority.Valid() {
		return nil, domain.ErrInvalidArgument
	}
	if strings.TrimSpace(in.Title) == "" {
		return nil, domain.ErrInvalidArgument
	}
	labels, e := parseIDs(in.LabelIDs, "labelIds")
	if e != nil {
		return nil, e
	}
	deps, e := s.dependencies(ctx, wid, tid, in.Dependencies)
	if e != nil {
		return nil, e
	}
	t := *old
	t.Title = strings.TrimSpace(in.Title)
	t.Description = strings.TrimSpace(in.Description)
	t.Status = in.Status
	t.Priority = in.Priority
	t.AssigneeID = strings.TrimSpace(in.AssigneeID)
	t.LabelIDs = labels
	t.Checklist = in.Checklist
	t.Dependencies = deps
	t.DueDate = in.DueDate
	t.StartDate = in.StartDate
	t.EstimateMinutes = in.EstimateMinutes
	t.Version++
	t.UpdatedAt = s.now().UTC()
	if e = validate(&t); e != nil {
		return nil, e
	}
	return s.repo.Apply(ctx, repository.Mutation{Task: &t, ExpectedVersion: old.Version, Event: event(&t, actor, "task.updated", t.UpdatedAt), Idempotency: idem(wid, actor, in.IdempotencyKey, h, t.UpdatedAt)})
}
func (s *Service) Delete(ctx context.Context, actor, w, id string, v int64, key string) error {
	_, e := s.deleted(ctx, actor, w, id, v, key, true)
	return e
}
func (s *Service) Restore(ctx context.Context, actor, w, id string, v int64, key string) (*domain.Task, error) {
	return s.deleted(ctx, actor, w, id, v, key, false)
}
func (s *Service) deleted(ctx context.Context, actor, w, id string, v int64, key string, del bool) (*domain.Task, error) {
	wid, e := parseID(w, "workspaceId")
	if e != nil {
		return nil, e
	}
	tid, e := parseID(id, "id")
	if e != nil {
		return nil, e
	}
	if e = s.authorize(ctx, wid, actor); e != nil {
		return nil, e
	}
	h := hash([]any{id, v, del})
	if replay, e := s.replay(ctx, wid, actor, key, h); e != nil || replay != nil {
		return replay, e
	}
	old, e := s.repo.Get(ctx, wid, tid, true)
	if e != nil {
		return nil, e
	}
	if old.Version != v {
		return nil, domain.ErrConflict
	}
	t := *old
	now := s.now().UTC()
	typ := "task.restored"
	if del {
		t.DeletedAt = &now
		t.DeletedBy = actor
		typ = "task.archived"
	} else {
		t.DeletedAt = nil
		t.DeletedBy = ""
	}
	t.Version++
	t.UpdatedAt = now
	return s.repo.Apply(ctx, repository.Mutation{Task: &t, ExpectedVersion: v, Event: event(&t, actor, typ, now), Idempotency: idem(wid, actor, key, h, now)})
}
func (s *Service) Reorder(ctx context.Context, actor, w, id string, v int64, before, after *float64, key string) (*domain.Task, error) {
	wid, e := parseID(w, "workspaceId")
	if e != nil {
		return nil, e
	}
	tid, e := parseID(id, "id")
	if e != nil {
		return nil, e
	}
	if e = s.authorize(ctx, wid, actor); e != nil {
		return nil, e
	}
	h := hash([]any{id, v, before, after})
	if replay, e := s.replay(ctx, wid, actor, key, h); e != nil || replay != nil {
		return replay, e
	}
	old, e := s.repo.Get(ctx, wid, tid, false)
	if e != nil {
		return nil, e
	}
	if old.Version != v {
		return nil, domain.ErrConflict
	}
	t := *old
	t.Position = domain.PositionBetween(before, after)
	t.Version++
	t.UpdatedAt = s.now().UTC()
	return s.repo.Apply(ctx, repository.Mutation{Task: &t, ExpectedVersion: v, Event: event(&t, actor, "task.reordered", t.UpdatedAt), Idempotency: idem(wid, actor, key, h, t.UpdatedAt)})
}
func (s *Service) List(ctx context.Context, actor string, in ListInput) (*ListResult, error) {
	wid, e := parseID(in.WorkspaceID, "workspaceId")
	if e != nil {
		return nil, e
	}
	if e = s.authorize(ctx, wid, actor); e != nil {
		return nil, e
	}
	f := repository.ListFilter{WorkspaceID: wid, Statuses: in.Statuses, Priorities: in.Priorities, AssigneeID: in.AssigneeID, ReporterID: in.ReporterID, Query: strings.TrimSpace(in.Query), IncludeDeleted: in.IncludeDeleted, DueFrom: in.DueFrom, DueTo: in.DueTo}
	if in.ProjectID != "" {
		x, e := parseID(in.ProjectID, "projectId")
		if e != nil {
			return nil, e
		}
		f.ProjectID = &x
	}
	if in.LabelID != "" {
		x, e := parseID(in.LabelID, "labelId")
		if e != nil {
			return nil, e
		}
		f.LabelID = &x
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	f.Limit = limit + 1
	if in.Cursor != "" {
		p, x, e := decodeCursor(in.Cursor)
		if e != nil {
			return nil, e
		}
		f.AfterPosition = &p
		f.AfterID = &x
	}
	tasks, e := s.repo.List(ctx, f)
	if e != nil {
		return nil, e
	}
	r := &ListResult{Tasks: tasks}
	if int64(len(tasks)) > limit {
		last := tasks[limit-1]
		r.Tasks = tasks[:limit]
		r.NextCursor = encodeCursor(last.Position, last.ID)
	}
	return r, nil
}
func (s *Service) BatchUpdate(ctx context.Context, actor, w, key string, items []BatchItem) ([]*domain.Task, error) {
	wid, e := parseID(w, "workspaceId")
	if e != nil {
		return nil, e
	}
	if e = s.authorize(ctx, wid, actor); e != nil {
		return nil, e
	}
	if len(items) == 0 || len(items) > 100 {
		return nil, domain.ErrInvalidArgument
	}
	h := hash(items)
	if key != "" {
		r, e := s.repo.FindIdempotency(ctx, wid, actor, key)
		if e == nil {
			if r.RequestHash != h {
				return nil, domain.ErrIdempotencyConflict
			}
			return r.Tasks, nil
		}
		if e != domain.ErrNotFound {
			return nil, e
		}
	}
	now := s.now().UTC()
	ms := make([]repository.Mutation, 0, len(items))
	for _, i := range items {
		tid, e := parseID(i.ID, "id")
		if e != nil {
			return nil, e
		}
		old, e := s.repo.Get(ctx, wid, tid, false)
		if e != nil {
			return nil, e
		}
		if old.Version != i.Version {
			return nil, domain.ErrConflict
		}
		t := *old
		if i.Status != "" {
			if !domain.CanTransition(t.Status, i.Status) {
				return nil, domain.ErrInvalidTransition
			}
			t.Status = i.Status
		}
		if i.Priority != "" {
			if !i.Priority.Valid() {
				return nil, domain.ErrInvalidArgument
			}
			t.Priority = i.Priority
		}
		if i.AssigneeID != "" {
			t.AssigneeID = i.AssigneeID
		}
		t.Version++
		t.UpdatedAt = now
		ms = append(ms, repository.Mutation{Task: &t, ExpectedVersion: old.Version, Event: event(&t, actor, "task.batch_updated", now)})
	}
	if key != "" {
		ms[0].Idempotency = idem(wid, actor, key, h, now)
	}
	return s.repo.ApplyBatch(ctx, ms)
}

func (s *Service) authorize(ctx context.Context, wid primitive.ObjectID, actor string) error {
	if e := s.identities.Validate(ctx, actor); e != nil {
		return e
	}
	ok, e := s.repo.IsMember(ctx, wid, actor)
	if e != nil {
		return e
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}
func (s *Service) replay(ctx context.Context, wid primitive.ObjectID, actor, key, h string) (*domain.Task, error) {
	if key == "" {
		return nil, nil
	}
	if len(key) > 200 {
		return nil, domain.ErrInvalidArgument
	}
	r, e := s.repo.FindIdempotency(ctx, wid, actor, key)
	if e == domain.ErrNotFound {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	if r.RequestHash != h {
		return nil, domain.ErrIdempotencyConflict
	}
	return r.Task, nil
}
func (s *Service) dependencies(ctx context.Context, wid, tid primitive.ObjectID, raw []string) ([]domain.TaskDependency, error) {
	out := make([]domain.TaskDependency, 0, len(raw))
	for _, v := range raw {
		id, e := parseID(v, "dependencyId")
		if e != nil {
			return nil, e
		}
		if id == tid {
			return nil, domain.ErrDependencyCycle
		}
		if tid != primitive.NilObjectID {
			yes, e := s.repo.DependencyPathExists(ctx, wid, id, tid)
			if e != nil {
				return nil, e
			}
			if yes {
				return nil, domain.ErrDependencyCycle
			}
		}
		out = append(out, domain.TaskDependency{TaskID: tid, DependsOnTaskID: id})
	}
	return out, nil
}
func parseID(v, f string) (primitive.ObjectID, error) {
	id, e := primitive.ObjectIDFromHex(strings.TrimSpace(v))
	if e != nil {
		return primitive.NilObjectID, fmt.Errorf("%w: %s", domain.ErrInvalidArgument, f)
	}
	return id, nil
}
func parseIDs(v []string, f string) ([]primitive.ObjectID, error) {
	out := make([]primitive.ObjectID, 0, len(v))
	for _, x := range v {
		id, e := parseID(x, f)
		if e != nil {
			return nil, e
		}
		out = append(out, id)
	}
	return out, nil
}
func validate(t *domain.Task) error {
	if t.EstimateMinutes < 0 {
		return domain.ErrInvalidArgument
	}
	if t.StartDate != nil && t.DueDate != nil && t.DueDate.Before(*t.StartDate) {
		return domain.ErrInvalidArgument
	}
	return nil
}
func hash(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func idem(w primitive.ObjectID, a, k, h string, n time.Time) *repository.Idempotency {
	if k == "" {
		return nil
	}
	return &repository.Idempotency{WorkspaceID: w, ActorID: a, Key: k, RequestHash: h, CreatedAt: n, ExpiresAt: n.Add(24 * time.Hour)}
}
func event(t *domain.Task, a, typ string, n time.Time) *domain.ActivityEvent {
	return &domain.ActivityEvent{ID: primitive.NewObjectID(), WorkspaceID: t.WorkspaceID, TaskID: t.ID, ActorID: a, Type: typ, Version: t.Version, CreatedAt: n}
}

type pageCursor struct {
	Position float64 `json:"p"`
	ID       string  `json:"i"`
}

func encodeCursor(p float64, id primitive.ObjectID) string {
	b, _ := json.Marshal(pageCursor{p, id.Hex()})
	return base64.RawURLEncoding.EncodeToString(b)
}
func decodeCursor(v string) (float64, primitive.ObjectID, error) {
	b, e := base64.RawURLEncoding.DecodeString(v)
	if e != nil {
		return 0, primitive.NilObjectID, domain.ErrInvalidArgument
	}
	var c pageCursor
	if json.Unmarshal(b, &c) != nil {
		return 0, primitive.NilObjectID, domain.ErrInvalidArgument
	}
	id, e := parseID(c.ID, "cursor")
	return c.Position, id, e
}
