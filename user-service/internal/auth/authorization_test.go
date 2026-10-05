package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
)

func TestRequireAuthenticatedUser(t *testing.T) {
	expected := &AuthenticatedUser{
		ID:    "user-123",
		Email: "john@example.com",
		Role:  domain.UserRoleUser,
	}

	ctx := WithAuthenticatedUser(context.Background(), expected)
	actual, err := RequireAuthenticatedUser(ctx)
	if err != nil {
		t.Fatalf("RequireAuthenticatedUser() error = %v", err)
	}

	if actual.ID != expected.ID {
		t.Errorf("ID = %s, want %s", actual.ID, expected.ID)
	}
}

func TestRequireAuthenticatedUserMissing(t *testing.T) {
	_, err := RequireAuthenticatedUser(context.Background())
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestRequireRoleAdmin(t *testing.T) {
	user := &AuthenticatedUser{
		ID:    "admin-123",
		Email: "admin@example.com",
		Role:  domain.UserRoleAdmin,
	}

	ctx := WithAuthenticatedUser(context.Background(), user)
	_, err := RequireRole(ctx, domain.UserRoleAdmin)
	if err != nil {
		t.Fatalf("RequireRole() error = %v", err)
	}
}

func TestRequireRoleForbidden(t *testing.T) {
	user := &AuthenticatedUser{
		ID:    "user-123",
		Email: "john@example.com",
		Role:  domain.UserRoleUser,
	}

	ctx := WithAuthenticatedUser(context.Background(), user)
	_, err := RequireRole(ctx, domain.UserRoleAdmin)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestRequireSelfOrAdminAllowsSelf(t *testing.T) {
	user := &AuthenticatedUser{
		ID:    "user-123",
		Email: "john@example.com",
		Role:  domain.UserRoleUser,
	}

	ctx := WithAuthenticatedUser(context.Background(), user)
	_, err := RequireSelfOrAdmin(ctx, user.ID)
	if err != nil {
		t.Fatalf("RequireSelfOrAdmin() error = %v", err)
	}
}

func TestRequireSelfOrAdminAllowsAdmin(t *testing.T) {
	user := &AuthenticatedUser{
		ID:    "admin-123",
		Email: "admin@example.com",
		Role:  domain.UserRoleAdmin,
	}

	ctx := WithAuthenticatedUser(context.Background(), user)
	_, err := RequireSelfOrAdmin(ctx, "user-123")
	if err != nil {
		t.Fatalf("RequireSelfOrAdmin() error = %v", err)
	}
}

func TestRequireSelfOrAdminForbidden(t *testing.T) {
	user := &AuthenticatedUser{
		ID:    "user-123",
		Email: "john@example.com",
		Role:  domain.UserRoleUser,
	}

	ctx := WithAuthenticatedUser(context.Background(), user)
	_, err := RequireSelfOrAdmin(ctx, "other-user")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
