package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService struct {
	userRepository repository.UserRespository
	jwtManager     *auth.JWTManager
}

func NewAuthService(userRepo repository.UserRespository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		userRepository: userRepo,
		jwtManager:     jwtManager,
	}
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	User        *domain.User
	AccessToken string
	ExpiresIn   int64
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.ToLower(strings.TrimSpace(input.Email))

	if err := ValidateUser(name, email); err != nil {
		return nil, err
	}

	if err := validatePassword(input.Password); err != nil {
		return nil, err
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	user := &domain.User{
		ID:           primitive.NewObjectID(),
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         domain.UserRoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:        user,
		AccessToken: accessToken,
		ExpiresIn:   int64(s.jwtManager.AccessTokenDuration()),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	if email == "" || input.Password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}

		return nil, err
	}

	if user.PasswordHash == "" {
		return nil, domain.ErrInvalidCredentials
	}

	if err := comparePassword(user.PasswordHash, input.Password); err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:        user,
		AccessToken: accessToken,
		ExpiresIn:   int64(s.jwtManager.AccessTokenDuration()),
	}, nil
}
