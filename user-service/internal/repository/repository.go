package repository

import (
	"context"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
)

type ListUserParams struct {
	Size  int
	Limit int
}

// ListUserParmas is kept as a compatibility alias for existing consumers.
type ListUserParmas = ListUserParams

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, params ListUserParams) ([]*domain.User, int64, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

// UserRespository is kept as a compatibility alias. Use UserRepository in new code.
type UserRespository = UserRepository
