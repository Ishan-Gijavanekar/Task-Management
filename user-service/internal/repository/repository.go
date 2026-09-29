package repository

import (
	"context"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
)

type ListUserParmas struct {
	Size  int
	Limit int
}

type UserRespository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, params ListUserParmas) ([]*domain.User, int64, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}
