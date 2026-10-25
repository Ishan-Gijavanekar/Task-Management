package repository

import (
	"context"
	"time"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Idempotency struct {
	WorkspaceID primitive.ObjectID `bson:"workspaceId"`
	ActorID     string             `bson:"actorId"`
	Key         string             `bson:"key"`
	RequestHash string             `bson:"requestHash"`
	Task        *domain.Task       `bson:"task,omitempty"`
	Tasks       []*domain.Task     `bson:"tasks,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt"`
	ExpiresAt   time.Time          `bson:"expiresAt"`
}

type Mutation struct {
	Task            *domain.Task
	ExpectedVersion int64
	Event           *domain.ActivityEvent
	Idempotency     *Idempotency
	Create          bool
}

type ListFilter struct {
	WorkspaceID    primitive.ObjectID
	ProjectID      *primitive.ObjectID
	ProjectIDs     []primitive.ObjectID
	Statuses       []domain.TaskStatus
	Priorities     []domain.Priority
	AssigneeID     string
	ReporterID     string
	LabelID        *primitive.ObjectID
	DueFrom        *time.Time
	DueTo          *time.Time
	Query          string
	IncludeDeleted bool
	Limit          int64
	AfterPosition  *float64
	AfterID        *primitive.ObjectID
}

type TaskRepository interface {
	Apply(context.Context, Mutation) (*domain.Task, error)
	ApplyBatch(context.Context, []Mutation) ([]*domain.Task, error)
	Get(context.Context, primitive.ObjectID, primitive.ObjectID, bool) (*domain.Task, error)
	List(context.Context, ListFilter) ([]*domain.Task, error)
	FindIdempotency(context.Context, primitive.ObjectID, string, string) (*Idempotency, error)
	GetMembership(context.Context, primitive.ObjectID, string) (*domain.Membership, error)
	GetWorkspace(context.Context, primitive.ObjectID, bool) (*domain.Workspace, error)
	GetProject(context.Context, primitive.ObjectID, primitive.ObjectID, bool) (*domain.Project, error)
	ListProjects(context.Context, primitive.ObjectID, bool) ([]*domain.Project, error)
	DependencyPathExists(context.Context, primitive.ObjectID, primitive.ObjectID, primitive.ObjectID) (bool, error)
	Ping(context.Context) error
}

type AdministrationRepository interface {
	CreateWorkspace(context.Context, *domain.Workspace, *domain.Membership, *domain.AuditEvent) error
	GetWorkspace(context.Context, primitive.ObjectID, bool) (*domain.Workspace, error)
	UpdateWorkspace(context.Context, *domain.Workspace, int64, *domain.AuditEvent) error
	TransferOwnership(context.Context, *domain.Workspace, int64, *domain.Membership, *domain.Membership, *domain.AuditEvent) error
	CreateProject(context.Context, *domain.Project, *domain.AuditEvent) error
	GetProject(context.Context, primitive.ObjectID, primitive.ObjectID, bool) (*domain.Project, error)
	ListProjects(context.Context, primitive.ObjectID, bool) ([]*domain.Project, error)
	UpdateProject(context.Context, *domain.Project, int64, *domain.AuditEvent) error
	GetMembership(context.Context, primitive.ObjectID, string) (*domain.Membership, error)
	ListMemberships(context.Context, primitive.ObjectID) ([]*domain.Membership, error)
	UpsertMembership(context.Context, *domain.Membership, *domain.AuditEvent) error
	RemoveMembership(context.Context, primitive.ObjectID, string, *domain.AuditEvent) error
	CreateInvitation(context.Context, *domain.Invitation, *domain.AuditEvent) error
	GetInvitation(context.Context, primitive.ObjectID, primitive.ObjectID) (*domain.Invitation, error)
	GetInvitationByTokenHash(context.Context, string) (*domain.Invitation, error)
	UpdateInvitation(context.Context, *domain.Invitation, int64, *domain.AuditEvent) error
	AcceptInvitation(context.Context, *domain.Invitation, int64, *domain.Membership, *domain.AuditEvent) error
	ConsumeInvitationQuota(context.Context, primitive.ObjectID, string, time.Time, int64) error
	ListAuditEvents(context.Context, primitive.ObjectID, int64) ([]*domain.AuditEvent, error)
}
