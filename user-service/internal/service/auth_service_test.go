package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
)

type fakeUserRepository struct {
	users map[string]*domain.User
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (r *fakeUserRepository) Create(
	ctx context.Context,
	user *domain.User,
) error {

	if _, exists := r.users[user.Email]; exists {
		return domain.ErrUserAlreadyExsists
	}

	r.users[user.Email] = user

	return nil
}

func (r *fakeUserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {

	user, exists := r.users[email]

	if !exists {
		return nil, domain.ErrUserNotFound
	}

	return user, nil
}

func (r *fakeUserRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (r *fakeUserRepository) List(
	ctx context.Context,
	params repository.ListUserParmas,
) ([]*domain.User, int64, error) {
	return nil, 0, nil
}

func (r *fakeUserRepository) Update(
	ctx context.Context,
	user *domain.User,
) error {
	return nil
}

func (r *fakeUserRepository) Delete(
	ctx context.Context,
	id string,
) error {
	return nil
}

func TestAuthServiceRegister(
	t *testing.T,
) {

	repo := newFakeUserRepository()

	jwtManager := auth.NewJWTManager(
		"test-secret",
		"user-service",
		15*time.Minute,
	)

	authService := NewAuthService(
		repo,
		jwtManager,
	)

	result, err := authService.Register(
		context.Background(),
		RegisterInput{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "Password123",
		},
	)

	if err != nil {
		t.Fatalf(
			"Register() error = %v",
			err,
		)
	}

	if result == nil {
		t.Fatal(
			"Register() returned nil result",
		)
	}

	if result.User == nil {
		t.Fatal(
			"Register() returned nil user",
		)
	}

	if result.User.Email != "john@example.com" {
		t.Errorf(
			"Email = %s, want john@example.com",
			result.User.Email,
		)
	}

	if result.User.PasswordHash == "" {
		t.Error(
			"PasswordHash should not be empty",
		)
	}

	if result.User.PasswordHash == "Password123" {
		t.Error(
			"password was stored in plain text",
		)
	}

	if result.User.Role != domain.UserRoleUser {
		t.Errorf(
			"Role = %s, want %s",
			result.User.Role,
			domain.UserRoleUser,
		)
	}

	if result.AccessToken == "" {
		t.Error(
			"AccessToken should not be empty",
		)
	}

	if result.ExpiresIn != 900 {
		t.Errorf(
			"ExpiresIn = %d, want 900",
			result.ExpiresIn,
		)
	}
}

func TestAuthServiceLogin(
	t *testing.T,
) {

	repo := newFakeUserRepository()

	jwtManager := auth.NewJWTManager(
		"test-secret",
		"user-service",
		15*time.Minute,
	)

	authService := NewAuthService(
		repo,
		jwtManager,
	)

	_, err := authService.Register(
		context.Background(),
		RegisterInput{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "Password123",
		},
	)

	if err != nil {
		t.Fatalf(
			"Register() error = %v",
			err,
		)
	}

	result, err := authService.Login(
		context.Background(),
		LoginInput{
			Email:    "john@example.com",
			Password: "Password123",
		},
	)

	if err != nil {
		t.Fatalf(
			"Login() error = %v",
			err,
		)
	}

	if result.AccessToken == "" {
		t.Error(
			"Login() returned empty access token",
		)
	}

	if result.User.Email != "john@example.com" {
		t.Errorf(
			"Email = %s, want john@example.com",
			result.User.Email,
		)
	}
}

func TestAuthServiceLoginWrongPassword(
	t *testing.T,
) {

	repo := newFakeUserRepository()

	jwtManager := auth.NewJWTManager(
		"test-secret",
		"user-service",
		15*time.Minute,
	)

	authService := NewAuthService(
		repo,
		jwtManager,
	)

	_, err := authService.Register(
		context.Background(),
		RegisterInput{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "Password123",
		},
	)

	if err != nil {
		t.Fatalf(
			"Register() error = %v",
			err,
		)
	}

	_, err = authService.Login(
		context.Background(),
		LoginInput{
			Email:    "john@example.com",
			Password: "WrongPassword",
		},
	)

	if !errors.Is(
		err,
		domain.ErrInvalidCredentials,
	) {

		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}
}

func TestAuthServiceLoginUnknownUser(
	t *testing.T,
) {

	repo := newFakeUserRepository()

	jwtManager := auth.NewJWTManager(
		"test-secret",
		"user-service",
		15*time.Minute,
	)

	authService := NewAuthService(
		repo,
		jwtManager,
	)

	_, err := authService.Login(
		context.Background(),
		LoginInput{
			Email:    "unknown@example.com",
			Password: "Password123",
		},
	)

	if !errors.Is(
		err,
		domain.ErrInvalidCredentials,
	) {

		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}
}

func TestAuthServiceDuplicateRegistration(
	t *testing.T,
) {

	repo := newFakeUserRepository()

	jwtManager := auth.NewJWTManager(
		"test-secret",
		"user-service",
		15*time.Minute,
	)

	authService := NewAuthService(
		repo,
		jwtManager,
	)

	input := RegisterInput{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "Password123",
	}

	_, err := authService.Register(
		context.Background(),
		input,
	)

	if err != nil {
		t.Fatalf(
			"first Register() error = %v",
			err,
		)
	}

	_, err = authService.Register(
		context.Background(),
		input,
	)

	if !errors.Is(
		err,
		domain.ErrUserAlreadyExsists,
	) {

		t.Fatalf(
			"expected ErrUserAlreadyExists, got %v",
			err,
		)
	}
}
