package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
	db          *mongo.Database
	tasks       *mongo.Collection
	activities  *mongo.Collection
	idempotency *mongo.Collection
	memberships *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) (*MongoRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("mongo database is required")
	}
	return &MongoRepository{db: db, tasks: db.Collection("tasks"), activities: db.Collection("task_activities"), idempotency: db.Collection("task_idempotency"), memberships: db.Collection("workspace_memberships")}, nil
}

func (r *MongoRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{Keys: bson.D{{Key: "workspaceId", Value: 1}, {Key: "projectId", Value: 1}, {Key: "status", Value: 1}, {Key: "position", Value: 1}, {Key: "_id", Value: 1}}},
		{Keys: bson.D{{Key: "workspaceId", Value: 1}, {Key: "assigneeId", Value: 1}, {Key: "deletedAt", Value: 1}, {Key: "updatedAt", Value: -1}}},
		{Keys: bson.D{{Key: "workspaceId", Value: 1}, {Key: "reporterId", Value: 1}, {Key: "updatedAt", Value: -1}}},
		{Keys: bson.D{{Key: "workspaceId", Value: 1}, {Key: "labelIds", Value: 1}, {Key: "updatedAt", Value: -1}}},
		{Keys: bson.D{{Key: "workspaceId", Value: 1}, {Key: "dueDate", Value: 1}}},
		{Keys: bson.D{{Key: "workspaceId", Value: 1}, {Key: "title", Value: "text"}, {Key: "description", Value: "text"}}},
	}
	if _, err := r.tasks.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("create task indexes: %w", err)
	}
	if _, err := r.activities.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "workspaceId", Value: 1}, {Key: "taskId", Value: 1}, {Key: "createdAt", Value: -1}}}); err != nil {
		return fmt.Errorf("create activity index: %w", err)
	}
	if _, err := r.memberships.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "workspaceId", Value: 1}, {Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true)}); err != nil {
		return fmt.Errorf("create membership index: %w", err)
	}
	_, err := r.idempotency.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "workspaceId", Value: 1}, {Key: "actorId", Value: 1}, {Key: "key", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
	})
	return err
}

func (r *MongoRepository) Apply(ctx context.Context, mutation Mutation) (*domain.Task, error) {
	var result *domain.Task
	session, err := r.db.Client().StartSession()
	if err != nil {
		return nil, fmt.Errorf("start transaction: %w", err)
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(tx mongo.SessionContext) (any, error) {
		if mutation.Idempotency != nil {
			existing, err := r.findIdempotency(tx, mutation.Idempotency.WorkspaceID, mutation.Idempotency.ActorID, mutation.Idempotency.Key)
			if err == nil {
				if existing.RequestHash != mutation.Idempotency.RequestHash {
					return nil, domain.ErrIdempotencyConflict
				}
				result = existing.Task
				return nil, nil
			}
			if !errors.Is(err, domain.ErrNotFound) {
				return nil, err
			}
		}
		if mutation.Create {
			if _, err := r.tasks.InsertOne(tx, mutation.Task); err != nil {
				return nil, mapMongoError(err)
			}
		} else {
			filter := bson.M{"_id": mutation.Task.ID, "workspaceId": mutation.Task.WorkspaceID, "version": mutation.ExpectedVersion}
			res, err := r.tasks.ReplaceOne(tx, filter, mutation.Task)
			if err != nil {
				return nil, mapMongoError(err)
			}
			if res.MatchedCount == 0 {
				return nil, domain.ErrConflict
			}
		}
		if _, err := r.activities.InsertOne(tx, mutation.Event); err != nil {
			return nil, fmt.Errorf("record activity: %w", err)
		}
		if mutation.Idempotency != nil {
			mutation.Idempotency.Task = mutation.Task
			if _, err := r.idempotency.InsertOne(tx, mutation.Idempotency); err != nil {
				return nil, mapMongoError(err)
			}
		}
		copy := *mutation.Task
		result = &copy
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *MongoRepository) ApplyBatch(ctx context.Context, mutations []Mutation) ([]*domain.Task, error) {
	results := make([]*domain.Task, 0, len(mutations))
	session, err := r.db.Client().StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(tx mongo.SessionContext) (any, error) {
		for _, mutation := range mutations {
			filter := bson.M{"_id": mutation.Task.ID, "workspaceId": mutation.Task.WorkspaceID, "version": mutation.ExpectedVersion}
			res, err := r.tasks.ReplaceOne(tx, filter, mutation.Task)
			if err != nil {
				return nil, mapMongoError(err)
			}
			if res.MatchedCount == 0 {
				return nil, domain.ErrConflict
			}
			if _, err := r.activities.InsertOne(tx, mutation.Event); err != nil {
				return nil, err
			}
			copy := *mutation.Task
			results = append(results, &copy)
		}
		if len(mutations) > 0 && mutations[0].Idempotency != nil {
			if _, err := r.idempotency.InsertOne(tx, mutations[0].Idempotency); err != nil {
				return nil, mapMongoError(err)
			}
		}
		return nil, nil
	})
	return results, err
}

func (r *MongoRepository) Get(ctx context.Context, workspaceID, id primitive.ObjectID, includeDeleted bool) (*domain.Task, error) {
	filter := bson.M{"_id": id, "workspaceId": workspaceID}
	if !includeDeleted {
		filter["deletedAt"] = bson.M{"$exists": false}
	}
	var task domain.Task
	if err := r.tasks.FindOne(ctx, filter).Decode(&task); err != nil {
		return nil, mapMongoError(err)
	}
	return &task, nil
}

func (r *MongoRepository) List(ctx context.Context, f ListFilter) ([]*domain.Task, error) {
	filter := bson.M{"workspaceId": f.WorkspaceID}
	if !f.IncludeDeleted {
		filter["deletedAt"] = bson.M{"$exists": false}
	}
	if f.ProjectID != nil {
		filter["projectId"] = *f.ProjectID
	}
	if len(f.Statuses) > 0 {
		filter["status"] = bson.M{"$in": f.Statuses}
	}
	if len(f.Priorities) > 0 {
		filter["priority"] = bson.M{"$in": f.Priorities}
	}
	if f.AssigneeID != "" {
		filter["assigneeId"] = f.AssigneeID
	}
	if f.ReporterID != "" {
		filter["reporterId"] = f.ReporterID
	}
	if f.LabelID != nil {
		filter["labelIds"] = *f.LabelID
	}
	if f.DueFrom != nil || f.DueTo != nil {
		rangeFilter := bson.M{}
		if f.DueFrom != nil {
			rangeFilter["$gte"] = *f.DueFrom
		}
		if f.DueTo != nil {
			rangeFilter["$lte"] = *f.DueTo
		}
		filter["dueDate"] = rangeFilter
	}
	if f.Query != "" {
		filter["$or"] = []bson.M{{"title": primitive.Regex{Pattern: regexp.QuoteMeta(f.Query), Options: "i"}}, {"description": primitive.Regex{Pattern: regexp.QuoteMeta(f.Query), Options: "i"}}}
	}
	if f.AfterPosition != nil && f.AfterID != nil {
		filter["$and"] = []bson.M{{"$or": []bson.M{{"position": bson.M{"$gt": *f.AfterPosition}}, {"position": *f.AfterPosition, "_id": bson.M{"$gt": *f.AfterID}}}}}
	}
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "_id", Value: 1}}).SetLimit(f.Limit)
	cursor, err := r.tasks.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var tasks []*domain.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *MongoRepository) FindIdempotency(ctx context.Context, workspaceID primitive.ObjectID, actorID, key string) (*Idempotency, error) {
	return r.findIdempotency(ctx, workspaceID, actorID, key)
}
func (r *MongoRepository) findIdempotency(ctx context.Context, workspaceID primitive.ObjectID, actorID, key string) (*Idempotency, error) {
	var record Idempotency
	err := r.idempotency.FindOne(ctx, bson.M{"workspaceId": workspaceID, "actorId": actorID, "key": key}).Decode(&record)
	if err != nil {
		return nil, mapMongoError(err)
	}
	return &record, nil
}

func (r *MongoRepository) IsMember(ctx context.Context, workspaceID primitive.ObjectID, userID string) (bool, error) {
	err := r.memberships.FindOne(ctx, bson.M{"workspaceId": workspaceID, "userId": userID}).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	return err == nil, err
}

func (r *MongoRepository) DependencyPathExists(ctx context.Context, workspaceID, from, target primitive.ObjectID) (bool, error) {
	seen := map[primitive.ObjectID]bool{}
	queue := []primitive.ObjectID{from}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if id == target {
			return true, nil
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		var task domain.Task
		err := r.tasks.FindOne(ctx, bson.M{"_id": id, "workspaceId": workspaceID}).Decode(&task)
		if errors.Is(err, mongo.ErrNoDocuments) {
			continue
		}
		if err != nil {
			return false, err
		}
		for _, dep := range task.Dependencies {
			queue = append(queue, dep.DependsOnTaskID)
		}
	}
	return false, nil
}

func (r *MongoRepository) Ping(ctx context.Context) error { return r.db.Client().Ping(ctx, nil) }

func mapMongoError(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.ErrNotFound
	}
	if mongo.IsDuplicateKeyError(err) {
		return domain.ErrConflict
	}
	return err
}
