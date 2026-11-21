package service

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"github.com/Ishan-Gijavanekar/task-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type memoryRepo struct {
	tasks         map[primitive.ObjectID]*domain.Task
	members       map[string]domain.WorkspaceRole
	projects      map[primitive.ObjectID]*domain.Project
	workspaces    map[primitive.ObjectID]*domain.Workspace
	keys          map[string]*repository.Idempotency
	events        int
	comments      map[primitive.ObjectID]*domain.Comment
	activities    []*domain.ActivityEvent
	notifications []*domain.Notification
	preferences   *domain.NotificationPreferences
}

func newMemory(w primitive.ObjectID, actor string) *memoryRepo {
	projectID, _ := primitive.ObjectIDFromHex("000000000000000000000001")
	return &memoryRepo{tasks: map[primitive.ObjectID]*domain.Task{}, members: map[string]domain.WorkspaceRole{w.Hex() + actor: domain.RoleMember}, projects: map[primitive.ObjectID]*domain.Project{projectID: {ID: projectID, WorkspaceID: w, Visibility: domain.ProjectWorkspace}}, workspaces: map[primitive.ObjectID]*domain.Workspace{w: {ID: w}}, keys: map[string]*repository.Idempotency{}, comments: map[primitive.ObjectID]*domain.Comment{}}
}
func (m *memoryRepo) Apply(_ context.Context, x repository.Mutation) (*domain.Task, error) {
	if x.Idempotency != nil {
		if old := m.keys[x.Idempotency.Key]; old != nil {
			if old.RequestHash != x.Idempotency.RequestHash {
				return nil, domain.ErrIdempotencyConflict
			}
			return old.Task, nil
		}
	}
	if !x.Create {
		old := m.tasks[x.Task.ID]
		if old == nil {
			return nil, domain.ErrNotFound
		}
		if old.Version != x.ExpectedVersion {
			return nil, domain.ErrConflict
		}
	}
	copy := *x.Task
	m.tasks[x.Task.ID] = &copy
	m.events++
	if x.Idempotency != nil {
		x.Idempotency.Task = &copy
		m.keys[x.Idempotency.Key] = x.Idempotency
	}
	return &copy, nil
}
func (m *memoryRepo) ApplyBatch(ctx context.Context, x []repository.Mutation) ([]*domain.Task, error) {
	out := make([]*domain.Task, 0, len(x))
	for _, v := range x {
		t, e := m.Apply(ctx, v)
		if e != nil {
			return nil, e
		}
		out = append(out, t)
	}
	if len(x) > 0 && x[0].Idempotency != nil {
		x[0].Idempotency.Tasks = out
		m.keys[x[0].Idempotency.Key] = x[0].Idempotency
	}
	return out, nil
}
func (m *memoryRepo) Get(_ context.Context, w, id primitive.ObjectID, deleted bool) (*domain.Task, error) {
	t := m.tasks[id]
	if t == nil || t.WorkspaceID != w || (!deleted && t.DeletedAt != nil) {
		return nil, domain.ErrNotFound
	}
	c := *t
	return &c, nil
}
func (m *memoryRepo) List(_ context.Context, f repository.ListFilter) ([]*domain.Task, error) {
	var out []*domain.Task
	for _, t := range m.tasks {
		if t.WorkspaceID != f.WorkspaceID {
			continue
		}
		if f.ProjectID != nil && t.ProjectID != *f.ProjectID {
			continue
		}
		if f.ProjectIDs != nil {
			allowed := false
			for _, id := range f.ProjectIDs {
				if t.ProjectID == id {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}
		if f.AfterPosition != nil && (t.Position < *f.AfterPosition || (t.Position == *f.AfterPosition && t.ID.Hex() <= f.AfterID.Hex())) {
			continue
		}
		c := *t
		out = append(out, &c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Position == out[j].Position {
			return out[i].ID.Hex() < out[j].ID.Hex()
		}
		return out[i].Position < out[j].Position
	})
	if int64(len(out)) > f.Limit {
		out = out[:f.Limit]
	}
	return out, nil
}
func (m *memoryRepo) FindIdempotency(_ context.Context, _ primitive.ObjectID, _, key string) (*repository.Idempotency, error) {
	r := m.keys[key]
	if r == nil {
		return nil, domain.ErrNotFound
	}
	return r, nil
}
func (m *memoryRepo) GetMembership(_ context.Context, w primitive.ObjectID, a string) (*domain.Membership, error) {
	role, ok := m.members[w.Hex()+a]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &domain.Membership{WorkspaceID: w, UserID: a, Role: role, Status: domain.MembershipActive}, nil
}
func (m *memoryRepo) LabelsExist(_ context.Context, _ primitive.ObjectID, _ []primitive.ObjectID) (bool, error) {
	return true, nil
}
func (m *memoryRepo) CreateComment(_ context.Context, comment *domain.Comment, activity *domain.ActivityEvent, _ *domain.OutboxEvent) error {
	m.comments[comment.ID] = comment
	m.activities = append(m.activities, activity)
	for _, userID := range comment.MentionIDs {
		m.notifications = append(m.notifications, &domain.Notification{ID: primitive.NewObjectID(), WorkspaceID: comment.WorkspaceID, TaskID: comment.TaskID, UserID: userID, ActorID: comment.AuthorID, Type: "comment.mentioned", Message: "You were mentioned in a task comment", CreatedAt: comment.CreatedAt})
	}
	return nil
}
func (m *memoryRepo) UpdateComment(_ context.Context, comment *domain.Comment, expected int64, activity *domain.ActivityEvent, _ *domain.OutboxEvent) error {
	old := m.comments[comment.ID]
	if old == nil || old.Version != expected {
		return domain.ErrConflict
	}
	m.comments[comment.ID] = comment
	m.activities = append(m.activities, activity)
	return nil
}
func (m *memoryRepo) GetComment(_ context.Context, workspaceID, taskID, commentID primitive.ObjectID) (*domain.Comment, error) {
	comment := m.comments[commentID]
	if comment == nil || comment.WorkspaceID != workspaceID || comment.TaskID != taskID {
		return nil, domain.ErrNotFound
	}
	copy := *comment
	return &copy, nil
}
func (m *memoryRepo) ListComments(_ context.Context, workspaceID, taskID primitive.ObjectID, _ *primitive.ObjectID, limit int64) ([]*domain.Comment, error) {
	var out []*domain.Comment
	for _, comment := range m.comments {
		if comment.WorkspaceID == workspaceID && comment.TaskID == taskID && comment.DeletedAt == nil {
			out = append(out, comment)
		}
	}
	if int64(len(out)) > limit {
		out = out[:limit]
	}
	return out, nil
}
func (m *memoryRepo) ListActivities(_ context.Context, workspaceID, taskID primitive.ObjectID, _ *primitive.ObjectID, limit int64) ([]*domain.ActivityEvent, error) {
	var out []*domain.ActivityEvent
	for _, activity := range m.activities {
		if activity.WorkspaceID == workspaceID && activity.TaskID == taskID {
			out = append(out, activity)
		}
	}
	if int64(len(out)) > limit {
		out = out[:limit]
	}
	return out, nil
}
func (m *memoryRepo) ListNotifications(_ context.Context, userID string, unread bool, limit int64) ([]*domain.Notification, error) {
	var out []*domain.Notification
	for _, item := range m.notifications {
		if item.UserID == userID && (!unread || item.ReadAt == nil) {
			out = append(out, item)
		}
	}
	if int64(len(out)) > limit {
		out = out[:limit]
	}
	return out, nil
}
func (m *memoryRepo) MarkNotificationRead(_ context.Context, userID string, id *primitive.ObjectID, now time.Time) error {
	for _, item := range m.notifications {
		if item.UserID == userID && (id == nil || item.ID == *id) {
			item.ReadAt = &now
		}
	}
	return nil
}
func (m *memoryRepo) GetNotificationPreferences(_ context.Context, userID string) (*domain.NotificationPreferences, error) {
	if m.preferences == nil || m.preferences.UserID != userID {
		return nil, domain.ErrNotFound
	}
	copy := *m.preferences
	return &copy, nil
}
func (m *memoryRepo) SetNotificationPreferences(_ context.Context, preferences *domain.NotificationPreferences) error {
	copy := *preferences
	m.preferences = &copy
	return nil
}
func (m *memoryRepo) GetWorkspace(_ context.Context, w primitive.ObjectID, _ bool) (*domain.Workspace, error) {
	x := m.workspaces[w]
	if x == nil {
		return nil, domain.ErrNotFound
	}
	c := *x
	return &c, nil
}
func (m *memoryRepo) GetProject(_ context.Context, w, id primitive.ObjectID, _ bool) (*domain.Project, error) {
	p := m.projects[id]
	if p == nil || p.WorkspaceID != w {
		return nil, domain.ErrNotFound
	}
	c := *p
	return &c, nil
}
func (m *memoryRepo) ListProjects(_ context.Context, w primitive.ObjectID, _ bool) ([]*domain.Project, error) {
	var out []*domain.Project
	for _, p := range m.projects {
		if p.WorkspaceID == w {
			c := *p
			out = append(out, &c)
		}
	}
	return out, nil
}
func (m *memoryRepo) DependencyPathExists(context.Context, primitive.ObjectID, primitive.ObjectID, primitive.ObjectID) (bool, error) {
	return false, nil
}
func (m *memoryRepo) Ping(context.Context) error { return nil }

func setup(t *testing.T) (*Service, *memoryRepo, primitive.ObjectID, string) {
	t.Helper()
	w := primitive.NewObjectID()
	a := "user-1"
	r := newMemory(w, a)
	s, e := New(r, StaticIdentityValidator{})
	if e != nil {
		t.Fatal(e)
	}
	return s, r, w, a
}
func create(t *testing.T, s *Service, w primitive.ObjectID, a, key string) *domain.Task {
	t.Helper()
	task, e := s.Create(context.Background(), a, CreateInput{WorkspaceID: w.Hex(), ProjectID: "000000000000000000000001", Title: "task", IdempotencyKey: key})
	if e != nil {
		t.Fatal(e)
	}
	return task
}

func TestCreateIsIdempotentAndRecordsActivity(t *testing.T) {
	s, r, w, a := setup(t)
	one := create(t, s, w, a, "create-1")
	two := create(t, s, w, a, "create-1")
	if one.ID != two.ID || r.events != 1 {
		t.Fatalf("idempotency failed: events=%d", r.events)
	}
}
func TestTenantIsolation(t *testing.T) {
	s, _, w, a := setup(t)
	task := create(t, s, w, a, "")
	other := primitive.NewObjectID()
	if _, e := s.Get(context.Background(), a, other.Hex(), task.ID.Hex(), false); !errors.Is(e, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", e)
	}
}
func TestTransitionAndConcurrentUpdate(t *testing.T) {
	s, _, w, a := setup(t)
	task := create(t, s, w, a, "")
	bad := UpdateInput{WorkspaceID: w.Hex(), ID: task.ID.Hex(), Title: task.Title, Status: domain.StatusDone, Version: 1}
	if _, e := s.Update(context.Background(), a, bad); !errors.Is(e, domain.ErrInvalidTransition) {
		t.Fatalf("got %v", e)
	}
	good := bad
	good.Status = domain.StatusTodo
	if _, e := s.Update(context.Background(), a, good); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Update(context.Background(), a, good); !errors.Is(e, domain.ErrConflict) {
		t.Fatalf("expected version conflict, got %v", e)
	}
}
func TestCursorPaginationDoesNotRepeatTasks(t *testing.T) {
	s, _, w, a := setup(t)
	for i := 0; i < 3; i++ {
		p := float64(i + 1)
		pid := primitive.NewObjectID()
		s.repo.(*memoryRepo).projects[pid] = &domain.Project{ID: pid, WorkspaceID: w, Visibility: domain.ProjectWorkspace}
		_, e := s.Create(context.Background(), a, CreateInput{WorkspaceID: w.Hex(), ProjectID: pid.Hex(), Title: "task", Position: &p})
		if e != nil {
			t.Fatal(e)
		}
	}
	first, e := s.List(context.Background(), a, ListInput{WorkspaceID: w.Hex(), Limit: 2})
	if e != nil {
		t.Fatal(e)
	}
	if len(first.Tasks) != 2 || first.NextCursor == "" {
		t.Fatal("expected first page")
	}
	second, e := s.List(context.Background(), a, ListInput{WorkspaceID: w.Hex(), Limit: 2, Cursor: first.NextCursor})
	if e != nil {
		t.Fatal(e)
	}
	if len(second.Tasks) != 1 || second.Tasks[0].ID == first.Tasks[1].ID {
		t.Fatal("cursor repeated a task")
	}
}

func TestProjectTenantAndVisibilityEnforcedForTasks(t *testing.T) {
	s, repo, w, actor := setup(t)
	ctx := context.Background()
	foreignProject := primitive.NewObjectID()
	repo.projects[foreignProject] = &domain.Project{ID: foreignProject, WorkspaceID: primitive.NewObjectID(), Visibility: domain.ProjectWorkspace}
	if _, err := s.Create(ctx, actor, CreateInput{WorkspaceID: w.Hex(), ProjectID: foreignProject.Hex(), Title: "cross tenant"}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("cross-workspace project accepted: %v", err)
	}
	privateProject := primitive.NewObjectID()
	repo.projects[privateProject] = &domain.Project{ID: privateProject, WorkspaceID: w, Visibility: domain.ProjectPrivate}
	repo.members[w.Hex()+actor] = domain.RoleOwner
	s.authorizer.Invalidate(w, actor)
	task, err := s.Create(ctx, actor, CreateInput{WorkspaceID: w.Hex(), ProjectID: privateProject.Hex(), Title: "private"})
	if err != nil {
		t.Fatal(err)
	}
	repo.members[w.Hex()+actor] = domain.RoleViewer
	s.authorizer.Invalidate(w, actor)
	if _, err = s.Get(ctx, actor, w.Hex(), task.ID.Hex(), false); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("viewer read private task: %v", err)
	}
	result, err := s.List(ctx, actor, ListInput{WorkspaceID: w.Hex()})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Tasks) != 0 {
		t.Fatal("private task leaked through list")
	}
	if _, err = s.Create(ctx, actor, CreateInput{WorkspaceID: w.Hex(), ProjectID: "000000000000000000000001", Title: "viewer write"}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("viewer wrote task: %v", err)
	}
}

func TestTaskWritesRequireActiveWorkspaceAndProject(t *testing.T) {
	s, repo, w, actor := setup(t)
	ctx := context.Background()
	projectID, _ := primitive.ObjectIDFromHex("000000000000000000000001")
	if _, err := s.Create(ctx, actor, CreateInput{WorkspaceID: w.Hex(), ProjectID: projectID.Hex(), Title: "bad assignee", AssigneeID: "outsider"}); !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("non-member assignee accepted: %v", err)
	}
	now := time.Now()
	repo.projects[projectID].ArchivedAt = &now
	if _, err := s.Create(ctx, actor, CreateInput{WorkspaceID: w.Hex(), ProjectID: projectID.Hex(), Title: "archived project"}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("write to archived project accepted: %v", err)
	}
	repo.projects[projectID].ArchivedAt = nil
	repo.workspaces[w].ArchivedAt = &now
	if _, err := s.Create(ctx, actor, CreateInput{WorkspaceID: w.Hex(), ProjectID: projectID.Hex(), Title: "archived workspace"}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("write to archived workspace accepted: %v", err)
	}
}
