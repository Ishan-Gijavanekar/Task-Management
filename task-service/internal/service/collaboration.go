package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"github.com/Ishan-Gijavanekar/task-service/internal/policy"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentPage struct {
	Comments   []*domain.Comment `json:"comments"`
	NextCursor string            `json:"nextCursor,omitempty"`
}

type ActivityPage struct {
	Activities []*domain.ActivityEvent `json:"activities"`
	NextCursor string                  `json:"nextCursor,omitempty"`
}

func (s *Service) CreateComment(ctx context.Context, actor, workspaceID, taskID, body string, mentions []string) (*domain.Comment, error) {
	wid, tid, err := s.collaborationAccess(ctx, actor, workspaceID, taskID, policy.TaskWrite)
	if err != nil {
		return nil, err
	}
	body = strings.TrimSpace(body)
	if body == "" || len(body) > 10000 {
		return nil, fmt.Errorf("%w: comment body", domain.ErrInvalidArgument)
	}
	mentions, err = s.validateMentions(ctx, wid, actor, mentions)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	comment := &domain.Comment{ID: primitive.NewObjectID(), WorkspaceID: wid, TaskID: tid, AuthorID: actor, Body: body, MentionIDs: mentions, Version: 1, CreatedAt: now, UpdatedAt: now}
	if err = s.repo.CreateComment(ctx, comment, collaborationEvent(wid, tid, actor, "comment.created", comment.ID, 1, now), collaborationOutbox(wid, tid, "comment.created", comment.ID, actor, mentions, now)); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *Service) UpdateComment(ctx context.Context, actor, workspaceID, taskID, commentID, body string, mentions []string, version int64) (*domain.Comment, error) {
	wid, tid, err := s.collaborationAccess(ctx, actor, workspaceID, taskID, policy.TaskRead)
	if err != nil {
		return nil, err
	}
	cid, err := parseID(commentID, "commentId")
	if err != nil {
		return nil, err
	}
	comment, err := s.repo.GetComment(ctx, wid, tid, cid)
	if err != nil {
		return nil, err
	}
	membership, err := s.authorizer.Membership(ctx, wid, actor)
	if err != nil || (comment.AuthorID != actor && membership.Role != domain.RoleOwner && membership.Role != domain.RoleAdmin) {
		return nil, domain.ErrForbidden
	}
	if comment.DeletedAt != nil || version < 1 || comment.Version != version {
		return nil, domain.ErrConflict
	}
	body = strings.TrimSpace(body)
	if body == "" || len(body) > 10000 {
		return nil, domain.ErrInvalidArgument
	}
	mentions, err = s.validateMentions(ctx, wid, actor, mentions)
	if err != nil {
		return nil, err
	}
	comment.Body, comment.MentionIDs, comment.UpdatedAt, comment.Version = body, mentions, s.now().UTC(), comment.Version+1
	if err = s.repo.UpdateComment(ctx, comment, version, collaborationEvent(wid, tid, actor, "comment.updated", cid, comment.Version, comment.UpdatedAt), collaborationOutbox(wid, tid, "comment.updated", cid, actor, mentions, comment.UpdatedAt)); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *Service) DeleteComment(ctx context.Context, actor, workspaceID, taskID, commentID string, version int64) error {
	wid, tid, err := s.collaborationAccess(ctx, actor, workspaceID, taskID, policy.TaskRead)
	if err != nil {
		return err
	}
	cid, err := parseID(commentID, "commentId")
	if err != nil {
		return err
	}
	comment, err := s.repo.GetComment(ctx, wid, tid, cid)
	if err != nil {
		return err
	}
	membership, err := s.authorizer.Membership(ctx, wid, actor)
	if err != nil || (comment.AuthorID != actor && membership.Role != domain.RoleOwner && membership.Role != domain.RoleAdmin) {
		return domain.ErrForbidden
	}
	if comment.DeletedAt != nil || comment.Version != version {
		return domain.ErrConflict
	}
	now := s.now().UTC()
	comment.DeletedAt, comment.UpdatedAt, comment.Version = &now, now, comment.Version+1
	return s.repo.UpdateComment(ctx, comment, version, collaborationEvent(wid, tid, actor, "comment.deleted", cid, comment.Version, now), collaborationOutbox(wid, tid, "comment.deleted", cid, actor, nil, now))
}

func (s *Service) ListComments(ctx context.Context, actor, workspaceID, taskID, cursor string, limit int64) (*CommentPage, error) {
	wid, tid, err := s.collaborationAccess(ctx, actor, workspaceID, taskID, policy.TaskRead)
	if err != nil {
		return nil, err
	}
	before, err := collaborationCursor(cursor)
	if err != nil {
		return nil, err
	}
	limit = pageLimit(limit)
	items, err := s.repo.ListComments(ctx, wid, tid, before, limit+1)
	if err != nil {
		return nil, err
	}
	page := &CommentPage{Comments: items}
	if int64(len(items)) > limit {
		page.Comments, page.NextCursor = items[:limit], items[limit-1].ID.Hex()
	}
	return page, nil
}

func (s *Service) ListActivities(ctx context.Context, actor, workspaceID, taskID, cursor string, limit int64) (*ActivityPage, error) {
	wid, tid, err := s.collaborationAccess(ctx, actor, workspaceID, taskID, policy.TaskRead)
	if err != nil {
		return nil, err
	}
	before, err := collaborationCursor(cursor)
	if err != nil {
		return nil, err
	}
	limit = pageLimit(limit)
	items, err := s.repo.ListActivities(ctx, wid, tid, before, limit+1)
	if err != nil {
		return nil, err
	}
	page := &ActivityPage{Activities: items}
	if int64(len(items)) > limit {
		page.Activities, page.NextCursor = items[:limit], items[limit-1].ID.Hex()
	}
	return page, nil
}

func (s *Service) collaborationAccess(ctx context.Context, actor, workspaceID, taskID string, action policy.Action) (primitive.ObjectID, primitive.ObjectID, error) {
	wid, err := parseID(workspaceID, "workspaceId")
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}
	tid, err := parseID(taskID, "taskId")
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}
	if err = s.authorize(ctx, wid, actor, action); err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}
	task, err := s.repo.Get(ctx, wid, tid, false)
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}
	if err = s.authorizeProject(ctx, wid, task.ProjectID, actor, action); err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}
	return wid, tid, nil
}

func (s *Service) validateMentions(ctx context.Context, wid primitive.ObjectID, actor string, mentions []string) ([]string, error) {
	unique := map[string]bool{}
	out := make([]string, 0, len(mentions))
	for _, id := range mentions {
		id = strings.TrimSpace(id)
		if id == "" || id == actor || unique[id] {
			continue
		}
		membership, err := s.authorizer.Membership(ctx, wid, id)
		if err != nil || membership.Status != domain.MembershipActive {
			return nil, fmt.Errorf("%w: mentioned users must be active workspace members", domain.ErrInvalidArgument)
		}
		unique[id] = true
		out = append(out, id)
	}
	return out, nil
}

func collaborationCursor(raw string) (*primitive.ObjectID, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	id, err := primitive.ObjectIDFromHex(raw)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}
	return &id, nil
}

func pageLimit(limit int64) int64 {
	if limit < 1 {
		return 30
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func collaborationEvent(wid, tid primitive.ObjectID, actor, kind string, resource primitive.ObjectID, version int64, now time.Time) *domain.ActivityEvent {
	return &domain.ActivityEvent{ID: primitive.NewObjectID(), WorkspaceID: wid, TaskID: tid, ActorID: actor, Type: kind, Version: version, Changes: map[string]any{"commentId": resource.Hex()}, CreatedAt: now}
}

func collaborationOutbox(wid, tid primitive.ObjectID, kind string, resource primitive.ObjectID, actor string, mentions []string, now time.Time) *domain.OutboxEvent {
	return &domain.OutboxEvent{ID: primitive.NewObjectID(), WorkspaceID: wid, TaskID: tid, Type: kind, Payload: map[string]any{"commentId": resource.Hex(), "actorId": actor, "mentionIds": mentions}, CreatedAt: now}
}
