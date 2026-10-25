package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *MongoRepository) transaction(ctx context.Context, fn func(mongo.SessionContext) error) error {
	session, err := r.db.Client().StartSession()
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(tx mongo.SessionContext) (any, error) { return nil, fn(tx) })
	return err
}

func insertAudit(ctx context.Context, collection *mongo.Collection, event *domain.AuditEvent) error {
	if event == nil {
		return nil
	}
	_, err := collection.InsertOne(ctx, event)
	return err
}

func (r *MongoRepository) CreateWorkspace(ctx context.Context, workspace *domain.Workspace, member *domain.Membership, audit *domain.AuditEvent) error {
	return r.transaction(ctx, func(tx mongo.SessionContext) error {
		if _, err := r.workspaces.InsertOne(tx, workspace); err != nil {
			return mapMongoError(err)
		}
		if _, err := r.memberships.InsertOne(tx, member); err != nil {
			return mapMongoError(err)
		}
		return insertAudit(tx, r.auditEvents, audit)
	})
}

func (r *MongoRepository) GetWorkspace(ctx context.Context, id primitive.ObjectID, includeArchived bool) (*domain.Workspace, error) {
	filter := bson.M{"_id": id}
	if !includeArchived {
		filter["archivedAt"] = bson.M{"$exists": false}
	}
	var workspace domain.Workspace
	if err := r.workspaces.FindOne(ctx, filter).Decode(&workspace); err != nil {
		return nil, mapMongoError(err)
	}
	return &workspace, nil
}

func (r *MongoRepository) UpdateWorkspace(ctx context.Context, workspace *domain.Workspace, expected int64, audit *domain.AuditEvent) error {
	return r.transaction(ctx, func(tx mongo.SessionContext) error {
		res, err := r.workspaces.ReplaceOne(tx, bson.M{"_id": workspace.ID, "version": expected}, workspace)
		if err != nil {
			return mapMongoError(err)
		}
		if res.MatchedCount == 0 {
			return domain.ErrConflict
		}
		return insertAudit(tx, r.auditEvents, audit)
	})
}

func (r *MongoRepository) TransferOwnership(ctx context.Context, workspace *domain.Workspace, expected int64, previousOwner, nextOwner *domain.Membership, audit *domain.AuditEvent) error {
	return r.transaction(ctx, func(tx mongo.SessionContext) error {
		res, err := r.workspaces.ReplaceOne(tx, bson.M{"_id": workspace.ID, "version": expected, "ownerId": previousOwner.UserID}, workspace)
		if err != nil {
			return mapMongoError(err)
		}
		if res.MatchedCount == 0 {
			return domain.ErrConflict
		}
		for _, member := range []*domain.Membership{previousOwner, nextOwner} {
			res, err = r.memberships.ReplaceOne(tx, bson.M{"workspaceId": member.WorkspaceID, "userId": member.UserID}, member)
			if err != nil {
				return mapMongoError(err)
			}
			if res.MatchedCount == 0 {
				return domain.ErrConflict
			}
		}
		return insertAudit(tx, r.auditEvents, audit)
	})
}

func (r *MongoRepository) CreateProject(ctx context.Context, project *domain.Project, audit *domain.AuditEvent) error {
	return r.transaction(ctx, func(tx mongo.SessionContext) error {
		if _, err := r.projects.InsertOne(tx, project); err != nil {
			return mapMongoError(err)
		}
		return insertAudit(tx, r.auditEvents, audit)
	})
}

func (r *MongoRepository) GetProject(ctx context.Context, workspaceID, id primitive.ObjectID, includeArchived bool) (*domain.Project, error) {
	filter := bson.M{"_id": id, "workspaceId": workspaceID}
	if !includeArchived {
		filter["archivedAt"] = bson.M{"$exists": false}
	}
	var project domain.Project
	if err := r.projects.FindOne(ctx, filter).Decode(&project); err != nil {
		return nil, mapMongoError(err)
	}
	return &project, nil
}

func (r *MongoRepository) ListProjects(ctx context.Context, workspaceID primitive.ObjectID, includeArchived bool) ([]*domain.Project, error) {
	filter := bson.M{"workspaceId": workspaceID}
	if !includeArchived {
		filter["archivedAt"] = bson.M{"$exists": false}
	}
	cursor, err := r.projects.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "name", Value: 1}, {Key: "_id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var projects []*domain.Project
	if err := cursor.All(ctx, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *MongoRepository) UpdateProject(ctx context.Context, project *domain.Project, expected int64, audit *domain.AuditEvent) error {
	return r.transaction(ctx, func(tx mongo.SessionContext) error {
		res, err := r.projects.ReplaceOne(tx, bson.M{"_id": project.ID, "workspaceId": project.WorkspaceID, "version": expected}, project)
		if err != nil {
			return mapMongoError(err)
		}
		if res.MatchedCount == 0 {
			return domain.ErrConflict
		}
		return insertAudit(tx, r.auditEvents, audit)
	})
}

func (r *MongoRepository) ListMemberships(ctx context.Context, workspaceID primitive.ObjectID) ([]*domain.Membership, error) {
	cursor, err := r.memberships.Find(ctx, bson.M{"workspaceId": workspaceID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var memberships []*domain.Membership
	if err := cursor.All(ctx, &memberships); err != nil {
		return nil, err
	}
	return memberships, nil
}

func (r *MongoRepository) UpsertMembership(ctx context.Context, member *domain.Membership, audit *domain.AuditEvent) error {
	return r.transaction(ctx, func(tx mongo.SessionContext) error {
		_, err := r.memberships.ReplaceOne(tx, bson.M{"workspaceId": member.WorkspaceID, "userId": member.UserID}, member, options.Replace().SetUpsert(true))
		if err != nil {
			return mapMongoError(err)
		}
		return insertAudit(tx, r.auditEvents, audit)
	})
}

func (r *MongoRepository) RemoveMembership(ctx context.Context, workspaceID primitive.ObjectID, userID string, audit *domain.AuditEvent) error {
	return r.transaction(ctx, func(tx mongo.SessionContext) error {
		res, err := r.memberships.DeleteOne(tx, bson.M{"workspaceId": workspaceID, "userId": userID})
		if err != nil {
			return err
		}
		if res.DeletedCount == 0 {
			return domain.ErrNotFound
		}
		return insertAudit(tx, r.auditEvents, audit)
	})
}

func (r *MongoRepository) CreateInvitation(ctx context.Context, invitation *domain.Invitation, audit *domain.AuditEvent) error {
	return r.transaction(ctx, func(tx mongo.SessionContext) error {
		if _, err := r.invitations.InsertOne(tx, invitation); err != nil {
			return mapMongoError(err)
		}
		return insertAudit(tx, r.auditEvents, audit)
	})
}

func (r *MongoRepository) GetInvitation(ctx context.Context, workspaceID, id primitive.ObjectID) (*domain.Invitation, error) {
	var invitation domain.Invitation
	if err := r.invitations.FindOne(ctx, bson.M{"_id": id, "workspaceId": workspaceID}).Decode(&invitation); err != nil {
		return nil, mapMongoError(err)
	}
	return &invitation, nil
}

func (r *MongoRepository) GetInvitationByTokenHash(ctx context.Context, hash string) (*domain.Invitation, error) {
	var invitation domain.Invitation
	if err := r.invitations.FindOne(ctx, bson.M{"tokenHash": hash}).Decode(&invitation); err != nil {
		return nil, mapMongoError(err)
	}
	return &invitation, nil
}

func (r *MongoRepository) UpdateInvitation(ctx context.Context, invitation *domain.Invitation, audit *domain.AuditEvent) error {
	return r.transaction(ctx, func(tx mongo.SessionContext) error {
		res, err := r.invitations.ReplaceOne(tx, bson.M{"_id": invitation.ID, "workspaceId": invitation.WorkspaceID, "acceptedAt": bson.M{"$exists": false}, "revokedAt": bson.M{"$exists": false}}, invitation)
		if err != nil {
			return mapMongoError(err)
		}
		if res.MatchedCount == 0 {
			return domain.ErrConflict
		}
		return insertAudit(tx, r.auditEvents, audit)
	})
}

func (r *MongoRepository) AcceptInvitation(ctx context.Context, invitation *domain.Invitation, member *domain.Membership, audit *domain.AuditEvent) error {
	return r.transaction(ctx, func(tx mongo.SessionContext) error {
		acceptedAt := *invitation.AcceptedAt
		res, err := r.invitations.ReplaceOne(tx, bson.M{"_id": invitation.ID, "workspaceId": invitation.WorkspaceID, "acceptedAt": bson.M{"$exists": false}, "revokedAt": bson.M{"$exists": false}, "expiresAt": bson.M{"$gt": acceptedAt}}, invitation)
		if err != nil {
			return mapMongoError(err)
		}
		if res.MatchedCount == 0 {
			return domain.ErrConflict
		}
		_, err = r.memberships.ReplaceOne(tx, bson.M{"workspaceId": member.WorkspaceID, "userId": member.UserID}, member, options.Replace().SetUpsert(true))
		if err != nil {
			return mapMongoError(err)
		}
		return insertAudit(tx, r.auditEvents, audit)
	})
}

func (r *MongoRepository) CountRecentInvitations(ctx context.Context, workspaceID primitive.ObjectID, actorID string, since time.Time) (int64, error) {
	return r.invitations.CountDocuments(ctx, bson.M{"workspaceId": workspaceID, "invitedBy": actorID, "createdAt": bson.M{"$gte": since}})
}

func (r *MongoRepository) ListAuditEvents(ctx context.Context, workspaceID primitive.ObjectID, limit int64) ([]*domain.AuditEvent, error) {
	cursor, err := r.auditEvents.Find(ctx, bson.M{"workspaceId": workspaceID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var events []*domain.AuditEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}
