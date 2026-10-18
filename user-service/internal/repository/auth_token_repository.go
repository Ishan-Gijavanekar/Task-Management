package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuthTokenRepository interface {
	Create(ctx context.Context, token *domain.AuthToken) error
	FindByHash(ctx context.Context, tokenHash string, tokenType domain.AuthTokenType) (*domain.AuthToken, error)
	Revoke(ctx context.Context, id, reason string, usedAt time.Time) error
	RevokeFamily(ctx context.Context, familyID, reason string, revokedAt time.Time) error
	RevokeUser(ctx context.Context, userID, reason string, revokedAt time.Time) error
}

func (r *MongoAuthTokenRepository) RevokeUser(ctx context.Context, userID, reason string, revokedAt time.Time) error {
	_, err := r.collection.UpdateMany(ctx,
		bson.M{"userId": userID, "revokedAt": nil},
		bson.M{"$set": bson.M{"revokedAt": revokedAt, "revocationReason": reason}},
	)
	if err != nil {
		return fmt.Errorf("revoke user tokens: %w", err)
	}
	return nil
}

type MongoAuthTokenRepository struct {
	collection *mongo.Collection
}

func NewMongoAuthTokenRepository(database *mongo.Database) *MongoAuthTokenRepository {
	return &MongoAuthTokenRepository{collection: database.Collection("auth_tokens")}
}

func (r *MongoAuthTokenRepository) Create(ctx context.Context, token *domain.AuthToken) error {
	if token == nil {
		return fmt.Errorf("create auth token: token is nil")
	}
	_, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		return fmt.Errorf("create auth token: %w", err)
	}
	return nil
}

func (r *MongoAuthTokenRepository) FindByHash(ctx context.Context, tokenHash string, tokenType domain.AuthTokenType) (*domain.AuthToken, error) {
	var token domain.AuthToken
	err := r.collection.FindOne(ctx, bson.M{"tokenHash": tokenHash, "type": tokenType}).Decode(&token)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("find auth token: %w", err)
	}
	return &token, nil
}

func (r *MongoAuthTokenRepository) Revoke(ctx context.Context, id, reason string, usedAt time.Time) error {
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id, "revokedAt": nil},
		bson.M{"$set": bson.M{"lastUsedAt": usedAt, "revokedAt": usedAt, "revocationReason": reason}},
	)
	if err != nil {
		return fmt.Errorf("revoke auth token: %w", err)
	}
	if result.MatchedCount == 0 {
		return domain.ErrTokenRevoked
	}
	return nil
}

func (r *MongoAuthTokenRepository) RevokeFamily(ctx context.Context, familyID, reason string, revokedAt time.Time) error {
	_, err := r.collection.UpdateMany(ctx,
		bson.M{"familyId": familyID, "revokedAt": nil},
		bson.M{"$set": bson.M{"revokedAt": revokedAt, "revocationReason": reason}},
	)
	if err != nil {
		return fmt.Errorf("revoke token family: %w", err)
	}
	return nil
}

func (r *MongoAuthTokenRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tokenHash", Value: 1}, {Key: "type", Value: 1}}, Options: options.Index().SetUnique(true).SetName("unique_token_hash_type")},
		{Keys: bson.D{{Key: "familyId", Value: 1}}, Options: options.Index().SetName("token_family")},
		{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0).SetName("expire_auth_tokens")},
	}
	if _, err := r.collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("create auth token indexes: %w", err)
	}
	return nil
}
