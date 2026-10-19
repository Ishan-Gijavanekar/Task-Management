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
	IsMember(context.Context, primitive.ObjectID, string) (bool, error)
	DependencyPathExists(context.Context, primitive.ObjectID, primitive.ObjectID, primitive.ObjectID) (bool, error)
	Ping(context.Context) error
}
