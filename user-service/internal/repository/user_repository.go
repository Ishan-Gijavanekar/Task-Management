package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(database *mongo.Database) *MongoUserRepository {
	return &MongoUserRepository{
		collection: database.Collection("users"),
	}
}

func (r *MongoUserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrUserAlreadyExsists
		}

		return err
	}

	return nil
}

func (r *MongoUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user *domain.User

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.collection.FindOne(
		ctx,
		bson.M{"_id": objectId},
	).Decode(&user)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrInvalidUserId
		}

		return nil, err
	}

	return user, err
}

func (r *MongoUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user *domain.User

	err := r.collection.FindOne(
		ctx,
		bson.M{"email": email},
	).Decode(&user)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}

func (r *MongoUserRepository) List(ctx context.Context, params ListUserParmas) ([]*domain.User, int64, error) {
	skip := (params.Limit - 1) * params.Size

	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(params.Limit)).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("Error find: %v", err)
	}

	defer cursor.Close(ctx)

	users := make([]*domain.User, 0)

	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, fmt.Errorf("Error fetching: %v", err)
	}

	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, fmt.Errorf("Error counting documents: %v", err)
	}

	return users, total, nil
}

func (r *MongoUserRepository) Update(ctx context.Context, user *domain.User, id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrInvalidUserId
	}

	_, err = r.collection.UpdateByID(ctx, objectId, user)
	if err != nil {
		return fmt.Errorf("Error in update: %v", err)
	}

	return nil
}

func (r *MongoUserRepository) Delete(ctx context.Context, id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrInvalidUserId
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		return fmt.Errorf("Error in delete: %v", err)
	}

	return nil
}

func (r *MongoUserRepository) EnsureIndexes(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: options.Index().SetUnique(true).SetName("unique_email"),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("Error creating the indexes")
	}

	return nil
}
