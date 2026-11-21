package service

import (
	"context"
	"testing"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCommentLifecycleRecordsActivity(t *testing.T) {
	w := primitive.NewObjectID()
	repo := newMemory(w, "author")
	repo.members[w.Hex()+"mentioned"] = domain.RoleMember
	service, err := New(repo, StaticIdentityValidator{})
	if err != nil {
		t.Fatal(err)
	}
	task := create(t, service, w, "author", "comment-task")
	comment, err := service.CreateComment(context.Background(), "author", w.Hex(), task.ID.Hex(), " Hello ", []string{"mentioned"})
	if err != nil || comment.Body != "Hello" || len(comment.MentionIDs) != 1 {
		t.Fatalf("unexpected comment: %#v, %v", comment, err)
	}
	notifications, err := service.ListNotifications(context.Background(), "mentioned", true, 10)
	if err != nil || len(notifications) != 1 {
		t.Fatalf("unexpected notifications: %#v, %v", notifications, err)
	}
	if err = service.MarkNotificationsRead(context.Background(), "mentioned", notifications[0].ID.Hex()); err != nil {
		t.Fatal(err)
	}
	comment, err = service.UpdateComment(context.Background(), "author", w.Hex(), task.ID.Hex(), comment.ID.Hex(), "Updated", nil, comment.Version)
	if err != nil || comment.Version != 2 {
		t.Fatalf("unexpected update: %#v, %v", comment, err)
	}
	activities, err := service.ListActivities(context.Background(), "author", w.Hex(), task.ID.Hex(), "", 10)
	if err != nil || len(activities.Activities) != 2 {
		t.Fatalf("unexpected activities: %#v, %v", activities, err)
	}
}
