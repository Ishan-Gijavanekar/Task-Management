package service

import (
	"context"
	"strings"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService struct {
	userRepository repository.UserRespository
}

func NewUserService(userRepository *repository.UserRespository) *UserService {
	return &UserService{
		userRepository: *userRepository,
	}
}

type CreateUserInput struct {
	Name  string
	Email string
}

type UpdateUserInput struct {
	Name  string
	Email string
}

type ListUserInput struct {
	Page  int64
	Limit int64
}

type ListUserResult struct {
	Users []*domain.User
	Page  int64
	Total int64
	Limit int64
}

func (s *UserService) Create(ctx context.Context, user CreateUserInput) (*domain.User, error) {
	name := strings.TrimSpace(user.Name)
	email := strings.ToLower(strings.TrimSpace(user.Email))

	createUser := &domain.User{
		Name:      name,
		Email:     email,
		ID:        primitive.NewObjectID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.userRepository.Create(ctx, createUser)
	if err != nil {
		return nil, err
	}

	return createUser, nil
}

func (u *UserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := u.userRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := u.userRepository.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserService) List(ctx context.Context, input ListUserInput) (*ListUserResult, error) {
	page := input.Page
	limit := input.Limit

	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	users, total, err := u.userRepository.List(
		ctx,
		repository.ListUserParmas{
			Size:  int(page),
			Limit: int(limit),
		},
	)

	if err != nil {
		return nil, err
	}

	return &ListUserResult{
		Users: users,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (u *UserService) Update(ctx context.Context, id string, input UpdateUserInput) (*domain.User, error) {
	user, err := u.userRepository.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrInvalidUserId
	}

	name := strings.TrimSpace(input.Name)
	email := strings.ToLower(strings.TrimSpace(input.Email))

	update := &domain.User{
		ID:        user.ID,
		Name:      name,
		Email:     email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: time.Now(),
	}

	err = u.userRepository.Update(ctx, update)
	if err != nil {
		return nil, err
	}

	return update, err
}

func (u *UserService) Delete(ctx context.Context, id string) error {
	err := u.userRepository.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
