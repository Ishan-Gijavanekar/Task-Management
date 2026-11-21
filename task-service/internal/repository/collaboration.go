package repository

import (
	"context"
	"time"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *MongoRepository) CreateComment(ctx context.Context, comment *domain.Comment, activity *domain.ActivityEvent, outbox *domain.OutboxEvent) error {
	return r.transaction(ctx, func(tx mongo.SessionContext) error {
		if _, err := r.comments.InsertOne(tx, comment); err != nil {
			return mapMongoError(err)
		}
		if _, err := r.activities.InsertOne(tx, activity); err != nil {
			return err
		}
		if err := r.createMentionNotifications(tx, comment); err != nil {
			return err
		}
		_, err := r.outbox.InsertOne(tx, outbox)
		return err
	})
}

func (r *MongoRepository) UpdateComment(ctx context.Context, comment *domain.Comment, expected int64, activity *domain.ActivityEvent, outbox *domain.OutboxEvent) error {
	return r.transaction(ctx, func(tx mongo.SessionContext) error {
		result, err := r.comments.ReplaceOne(tx, bson.M{"_id": comment.ID, "workspaceId": comment.WorkspaceID, "taskId": comment.TaskID, "version": expected}, comment)
		if err != nil {
			return err
		}
		if result.MatchedCount == 0 {
			return domain.ErrConflict
		}
		if _, err = r.activities.InsertOne(tx, activity); err != nil {
			return err
		}
		if err = r.createMentionNotifications(tx, comment); err != nil {
			return err
		}
		_, err = r.outbox.InsertOne(tx, outbox)
		return err
	})
}

func (r *MongoRepository) createMentionNotifications(ctx context.Context, comment *domain.Comment) error {
	for _, userID := range comment.MentionIDs {
		var preferences domain.NotificationPreferences
		err := r.preferences.FindOne(ctx, bson.M{"_id": userID}).Decode(&preferences)
		if err == nil && !preferences.Mentions {
			continue
		}
		if err != nil && err != mongo.ErrNoDocuments {
			return err
		}
		n := &domain.Notification{ID: primitive.NewObjectID(), WorkspaceID: comment.WorkspaceID, TaskID: comment.TaskID, UserID: userID, ActorID: comment.AuthorID, Type: "comment.mentioned", Message: "You were mentioned in a task comment", SourceKey: comment.ID.Hex() + ":" + userID, CreatedAt: comment.UpdatedAt}
		_, err = r.notifications.UpdateOne(ctx, bson.M{"sourceKey": n.SourceKey}, bson.M{"$setOnInsert": n}, options.Update().SetUpsert(true))
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MongoRepository) ListNotifications(ctx context.Context, userID string, unread bool, limit int64) ([]*domain.Notification, error) {
	filter := bson.M{"userId": userID}
	if unread {
		filter["readAt"] = bson.M{"$exists": false}
	}
	cursor, err := r.notifications.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var items []*domain.Notification
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MongoRepository) MarkNotificationRead(ctx context.Context, userID string, id *primitive.ObjectID, now time.Time) error {
	filter := bson.M{"userId": userID, "readAt": bson.M{"$exists": false}}
	if id != nil {
		filter["_id"] = *id
	}
	_, err := r.notifications.UpdateMany(ctx, filter, bson.M{"$set": bson.M{"readAt": now}})
	return err
}

func (r *MongoRepository) GetNotificationPreferences(ctx context.Context, userID string) (*domain.NotificationPreferences, error) {
	var preferences domain.NotificationPreferences
	if err := r.preferences.FindOne(ctx, bson.M{"_id": userID}).Decode(&preferences); err != nil {
		return nil, mapMongoError(err)
	}
	return &preferences, nil
}

func (r *MongoRepository) SetNotificationPreferences(ctx context.Context, preferences *domain.NotificationPreferences) error {
	_, err := r.preferences.ReplaceOne(ctx, bson.M{"_id": preferences.UserID}, preferences, options.Replace().SetUpsert(true))
	return err
}

func (r *MongoRepository) GetComment(ctx context.Context, workspaceID, taskID, commentID primitive.ObjectID) (*domain.Comment, error) {
	var comment domain.Comment
	err := r.comments.FindOne(ctx, bson.M{"_id": commentID, "workspaceId": workspaceID, "taskId": taskID}).Decode(&comment)
	if err != nil {
		return nil, mapMongoError(err)
	}
	return &comment, nil
}

func (r *MongoRepository) ListComments(ctx context.Context, workspaceID, taskID primitive.ObjectID, before *primitive.ObjectID, limit int64) ([]*domain.Comment, error) {
	filter := bson.M{"workspaceId": workspaceID, "taskId": taskID, "deletedAt": bson.M{"$exists": false}}
	if before != nil {
		filter["_id"] = bson.M{"$lt": *before}
	}
	cursor, err := r.comments.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var comments []*domain.Comment
	if err = cursor.All(ctx, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *MongoRepository) ListActivities(ctx context.Context, workspaceID, taskID primitive.ObjectID, before *primitive.ObjectID, limit int64) ([]*domain.ActivityEvent, error) {
	filter := bson.M{"workspaceId": workspaceID, "taskId": taskID}
	if before != nil {
		filter["_id"] = bson.M{"$lt": *before}
	}
	cursor, err := r.activities.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var events []*domain.ActivityEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}
