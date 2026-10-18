package repository

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/database"
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMongoUserRepositoryIntegration(t *testing.T) {
	uri := os.Getenv("MONGO_INTEGRATION_URI")
	if uri == "" {
		t.Skip("set MONGO_INTEGRATION_URI to run MongoDB integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	databaseName := "user_service_test_" + primitive.NewObjectID().Hex()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mongoDB, err := database.NewMongoDB(uri, databaseName, *logger)
	if err != nil {
		t.Fatalf("connect to MongoDB: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_ = mongoDB.Database.Drop(cleanupCtx)
		_ = mongoDB.Close(cleanupCtx)
	})

	repo := NewMongoUserRepository(mongoDB.Database)
	if err := repo.EnsureIndexes(ctx); err != nil {
		t.Fatalf("EnsureIndexes() error: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Millisecond)
	user := &domain.User{
		ID: primitive.NewObjectID(), Name: "Ada Lovelace", Email: "ada@example.com",
		PasswordHash: "hash", Role: domain.UserRoleUser, CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	duplicate := *user
	duplicate.ID = primitive.NewObjectID()
	if err := repo.Create(ctx, &duplicate); !errors.Is(err, domain.ErrUserAlreadyExsists) {
		t.Fatalf("duplicate Create() error = %v, want ErrUserAlreadyExsists", err)
	}

	found, err := repo.GetByEmail(ctx, user.Email)
	if err != nil || found.ID != user.ID {
		t.Fatalf("GetByEmail() = %#v, %v", found, err)
	}
	found.Name = "Ada Byron"
	if err := repo.Update(ctx, found); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	users, total, err := repo.List(ctx, ListUserParmas{Size: 1, Limit: 10})
	if err != nil || total != 1 || len(users) != 1 || users[0].Name != "Ada Byron" {
		t.Fatalf("List() = %d users, total %d, error %v", len(users), total, err)
	}
	if err := repo.Delete(ctx, user.ID.Hex()); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if _, err := repo.GetByID(ctx, user.ID.Hex()); !errors.Is(err, domain.ErrInvalidUserId) {
		t.Fatalf("GetByID() after delete error = %v, want ErrInvalidUserId", err)
	}
}
