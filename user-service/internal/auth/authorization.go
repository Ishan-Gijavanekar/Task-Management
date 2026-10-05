package auth

import (
	"context"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
)

func RequireAuthenticatedUser(ctx context.Context) (*AuthenticatedUser, error) {
	user, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	return user, nil
}

func RequireRole(ctx context.Context, roles ...domain.UserRole) (*AuthenticatedUser, error) {
	user, err := RequireAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		if user.Role == role {
			return user, nil
		}
	}

	return nil, domain.ErrForbidden
}

func RequireSelfOrAdmin(ctx context.Context, targetUserID string) (*AuthenticatedUser, error) {
	user, err := RequireAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	if user.Role == domain.UserRoleAdmin || user.ID == targetUserID {
		return user, nil
	}

	return nil, domain.ErrForbidden
}
