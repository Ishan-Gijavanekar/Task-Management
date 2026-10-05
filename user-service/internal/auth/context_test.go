package auth

import (
	"context"
	"testing"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
)

func TestAuthenticatedUserContext(t *testing.T) {
	expected := &AuthenticatedUser{
		ID:    "user-123",
		Email: "john@example.com",
		Role:  domain.UserRoleUser,
	}

	ctx := WithAuthenticatedUser(context.Background(), expected)
	actual, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		t.Fatal("expected authenticated user in context")
	}

	if actual.ID != expected.ID {
		t.Errorf("ID = %s, want %s", actual.ID, expected.ID)
	}
	if actual.Email != expected.Email {
		t.Errorf("Email = %s, want %s", actual.Email, expected.Email)
	}
	if actual.Role != expected.Role {
		t.Errorf("Role = %s, want %s", actual.Role, expected.Role)
	}
}

func TestAuthenticatedUserMissingFromContext(t *testing.T) {
	_, ok := AuthenticatedUserFromContext(context.Background())
	if ok {
		t.Fatal("expected no authenticated user")
	}
}

func TestAuthenticatedUserNilFromContext(t *testing.T) {
	ctx := WithAuthenticatedUser(context.Background(), nil)
	_, ok := AuthenticatedUserFromContext(ctx)
	if ok {
		t.Fatal("expected nil authenticated user to be ignored")
	}
}
