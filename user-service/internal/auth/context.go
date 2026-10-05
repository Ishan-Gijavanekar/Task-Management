package auth

import (
	"context"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
)

type AuthenticatedUser struct {
	ID    string
	Email string
	Role  domain.UserRole
}

type contextKey string

const authenticatedUserContextKey contextKey = "authenticated-user"

func WithAuthenticatedUser(ctx context.Context, user *AuthenticatedUser) context.Context {
	return context.WithValue(ctx, authenticatedUserContextKey, user)
}

func AuthenticatedUserFromContext(ctx context.Context) (*AuthenticatedUser, bool) {
	user, ok := ctx.Value(authenticatedUserContextKey).(*AuthenticatedUser)
	if !ok || user == nil {
		return nil, false
	}

	return user, true
}
