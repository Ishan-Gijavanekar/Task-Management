package auth

import (
	"testing"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestJWTSigningKeyRotation(t *testing.T) {
	user := &domain.User{ID: primitive.NewObjectID(), Email: "user@example.com", Role: domain.UserRoleUser}
	oldManager := NewRotatingJWTManager("old", map[string]string{"old": "old-secret"}, "issuer", time.Minute)
	oldToken, err := oldManager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("generate old token: %v", err)
	}

	rotated := NewRotatingJWTManager("new", map[string]string{
		"old": "old-secret",
		"new": "new-secret",
	}, "issuer", time.Minute)
	if _, err := rotated.ValidateAccessToken(oldToken); err != nil {
		t.Fatalf("rotated manager rejected previous key: %v", err)
	}
	newToken, err := rotated.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("generate new token: %v", err)
	}
	if _, err := oldManager.ValidateAccessToken(newToken); err == nil {
		t.Fatal("old manager accepted token signed with new key")
	}
}
