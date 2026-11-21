package repository

import (
	"context"

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
		_, err = r.outbox.InsertOne(tx, outbox)
		return err
	})
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
