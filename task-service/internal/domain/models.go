package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkspaceRole string

const (
	RoleOwner  WorkspaceRole = "owner"
	RoleAdmin  WorkspaceRole = "admin"
	RoleMember WorkspaceRole = "member"
	RoleViewer WorkspaceRole = "viewer"
)

type Workspace struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name       string             `json:"name" bson:"name"`
	OwnerID    string             `json:"ownerId" bson:"ownerId"`
	Version    int64              `json:"version" bson:"version"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt" bson:"updatedAt"`
	ArchivedAt *time.Time         `json:"archivedAt,omitempty" bson:"archivedAt,omitempty"`
}

type MembershipStatus string

const (
	MembershipActive    MembershipStatus = "active"
	MembershipSuspended MembershipStatus = "suspended"
)

type Membership struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	WorkspaceID primitive.ObjectID `json:"workspaceId" bson:"workspaceId"`
	UserID      string             `json:"userId" bson:"userId"`
	Role        WorkspaceRole      `json:"role" bson:"role"`
	Status      MembershipStatus   `json:"status" bson:"status"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

func (r WorkspaceRole) Valid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember, RoleViewer:
		return true
	}
	return false
}

type ProjectVisibility string

const (
	ProjectWorkspace ProjectVisibility = "workspace"
	ProjectPrivate   ProjectVisibility = "private"
)

type WorkflowStatus struct {
	Key  string `json:"key" bson:"key"`
	Name string `json:"name" bson:"name"`
}

type Project struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	WorkspaceID primitive.ObjectID `json:"workspaceId" bson:"workspaceId"`
	Name        string             `json:"name" bson:"name"`
	Key         string             `json:"key" bson:"key"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Workflow    []WorkflowStatus   `json:"workflow" bson:"workflow"`
	Visibility  ProjectVisibility  `json:"visibility" bson:"visibility"`
	Version     int64              `json:"version" bson:"version"`
	ArchivedAt  *time.Time         `json:"archivedAt,omitempty" bson:"archivedAt,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type Invitation struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	WorkspaceID primitive.ObjectID `json:"workspaceId" bson:"workspaceId"`
	Email       string             `json:"email" bson:"email"`
	Role        WorkspaceRole      `json:"role" bson:"role"`
	TokenHash   string             `json:"-" bson:"tokenHash"`
	InvitedBy   string             `json:"invitedBy" bson:"invitedBy"`
	ExpiresAt   time.Time          `json:"expiresAt" bson:"expiresAt"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	AcceptedAt  *time.Time         `json:"acceptedAt,omitempty" bson:"acceptedAt,omitempty"`
	AcceptedBy  string             `json:"acceptedBy,omitempty" bson:"acceptedBy,omitempty"`
	RevokedAt   *time.Time         `json:"revokedAt,omitempty" bson:"revokedAt,omitempty"`
	RevokedBy   string             `json:"revokedBy,omitempty" bson:"revokedBy,omitempty"`
}

type AuditEvent struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	WorkspaceID primitive.ObjectID `json:"workspaceId" bson:"workspaceId"`
	ActorID     string             `json:"actorId" bson:"actorId"`
	Action      string             `json:"action" bson:"action"`
	Resource    string             `json:"resource" bson:"resource"`
	ResourceID  string             `json:"resourceId" bson:"resourceId"`
	Details     map[string]any     `json:"details,omitempty" bson:"details,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
}

type Label struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	WorkspaceID primitive.ObjectID `json:"workspaceId" bson:"workspaceId"`
	Name        string             `json:"name" bson:"name"`
	Color       string             `json:"color" bson:"color"`
}

type ChecklistItem struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Text      string             `json:"text" bson:"text"`
	Completed bool               `json:"completed" bson:"completed"`
	Position  float64            `json:"position" bson:"position"`
}

type TaskDependency struct {
	TaskID          primitive.ObjectID `json:"taskId" bson:"taskId"`
	DependsOnTaskID primitive.ObjectID `json:"dependsOnTaskId" bson:"dependsOnTaskId"`
}

type TaskStatus string

const (
	StatusBacklog    TaskStatus = "backlog"
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusBlocked    TaskStatus = "blocked"
	StatusDone       TaskStatus = "done"
	StatusCancelled  TaskStatus = "cancelled"
)

func (s TaskStatus) Valid() bool {
	switch s {
	case StatusBacklog, StatusTodo, StatusInProgress, StatusBlocked, StatusDone, StatusCancelled:
		return true
	}
	return false
}

var transitions = map[TaskStatus]map[TaskStatus]bool{
	StatusBacklog:    {StatusTodo: true, StatusCancelled: true},
	StatusTodo:       {StatusBacklog: true, StatusInProgress: true, StatusBlocked: true, StatusCancelled: true},
	StatusInProgress: {StatusTodo: true, StatusBlocked: true, StatusDone: true, StatusCancelled: true},
	StatusBlocked:    {StatusTodo: true, StatusInProgress: true, StatusCancelled: true},
	StatusDone:       {StatusInProgress: true},
	StatusCancelled:  {StatusBacklog: true},
}

func CanTransition(from, to TaskStatus) bool { return from == to || transitions[from][to] }

type Priority string

const (
	PriorityNone   Priority = "none"
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

func (p Priority) Valid() bool {
	switch p {
	case PriorityNone, PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	}
	return false
}

type Task struct {
	ID              primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	WorkspaceID     primitive.ObjectID   `json:"workspaceId" bson:"workspaceId"`
	ProjectID       primitive.ObjectID   `json:"projectId" bson:"projectId"`
	Title           string               `json:"title" bson:"title"`
	Description     string               `json:"description,omitempty" bson:"description,omitempty"`
	Status          TaskStatus           `json:"status" bson:"status"`
	Priority        Priority             `json:"priority" bson:"priority"`
	AssigneeID      string               `json:"assigneeId,omitempty" bson:"assigneeId,omitempty"`
	ReporterID      string               `json:"reporterId" bson:"reporterId"`
	LabelIDs        []primitive.ObjectID `json:"labelIds,omitempty" bson:"labelIds,omitempty"`
	Checklist       []ChecklistItem      `json:"checklist,omitempty" bson:"checklist,omitempty"`
	Dependencies    []TaskDependency     `json:"dependencies,omitempty" bson:"dependencies,omitempty"`
	DueDate         *time.Time           `json:"dueDate,omitempty" bson:"dueDate,omitempty"`
	StartDate       *time.Time           `json:"startDate,omitempty" bson:"startDate,omitempty"`
	EstimateMinutes int64                `json:"estimateMinutes,omitempty" bson:"estimateMinutes,omitempty"`
	Position        float64              `json:"position" bson:"position"`
	Version         int64                `json:"version" bson:"version"`
	CreatedAt       time.Time            `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt" bson:"updatedAt"`
	DeletedAt       *time.Time           `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	DeletedBy       string               `json:"deletedBy,omitempty" bson:"deletedBy,omitempty"`
}

func PositionBetween(before, after *float64) float64 {
	switch {
	case before == nil && after == nil:
		return 1024
	case before == nil:
		return *after / 2
	case after == nil:
		return *before + 1024
	default:
		return (*before + *after) / 2
	}
}

type ActivityEvent struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	WorkspaceID primitive.ObjectID `json:"workspaceId" bson:"workspaceId"`
	TaskID      primitive.ObjectID `json:"taskId" bson:"taskId"`
	ActorID     string             `json:"actorId" bson:"actorId"`
	Type        string             `json:"type" bson:"type"`
	Version     int64              `json:"version" bson:"version"`
	Changes     map[string]any     `json:"changes,omitempty" bson:"changes,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
}
