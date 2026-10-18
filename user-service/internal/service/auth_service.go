package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService struct {
	userRepository           repository.UserRespository
	tokenRepository          repository.AuthTokenRepository
	jwtManager               *auth.JWTManager
	refreshTokenDuration     time.Duration
	requireEmailVerification bool
	notifier                 AuthNotificationSender
}

type AuthServiceOption func(*AuthService)

func WithAuthTokenRepository(tokenRepository repository.AuthTokenRepository) AuthServiceOption {
	return func(service *AuthService) { service.tokenRepository = tokenRepository }
}

func WithRefreshTokenDuration(duration time.Duration) AuthServiceOption {
	return func(service *AuthService) { service.refreshTokenDuration = duration }
}

func WithEmailVerificationRequired(required bool) AuthServiceOption {
	return func(service *AuthService) { service.requireEmailVerification = required }
}

func WithAuthNotificationSender(notifier AuthNotificationSender) AuthServiceOption {
	return func(service *AuthService) { service.notifier = notifier }
}

func NewAuthService(userRepo repository.UserRespository, jwtManager *auth.JWTManager, options ...AuthServiceOption) *AuthService {
	service := &AuthService{
		userRepository:       userRepo,
		jwtManager:           jwtManager,
		refreshTokenDuration: 30 * 24 * time.Hour,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *AuthService) Validate() error {
	if s == nil || s.userRepository == nil {
		return fmt.Errorf("auth service: %w: user repository is required", domain.ErrInvalidConfiguration)
	}
	if s.jwtManager == nil {
		return fmt.Errorf("auth service: %w: JWT manager is required", domain.ErrInvalidConfiguration)
	}
	if s.tokenRepository == nil {
		return fmt.Errorf("auth service: %w: token repository is required", domain.ErrInvalidConfiguration)
	}
	if s.refreshTokenDuration <= 0 {
		return fmt.Errorf("auth service: %w: refresh token duration must be positive", domain.ErrInvalidConfiguration)
	}
	if s.requireEmailVerification && s.notifier == nil {
		return fmt.Errorf("auth service: %w: notification sender is required", domain.ErrInvalidConfiguration)
	}
	return nil
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email     string
	Password  string
	Device    string
	IPAddress string
}

type AuthResult struct {
	User             *domain.User
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int64
	RefreshExpiresIn int64
}

type RefreshInput struct {
	RefreshToken string
	Device       string
	IPAddress    string
}

func randomToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("generate random token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(bytes)
	sum := sha256.Sum256([]byte(raw))
	return raw, hex.EncodeToString(sum[:]), nil
}

func tokenHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func randomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *AuthService) issueSession(ctx context.Context, user *domain.User, familyID, device, ipAddress string) (*AuthResult, error) {
	if err := ensureAccountActive(user); err != nil {
		return nil, err
	}
	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}
	if s.tokenRepository == nil {
		return nil, fmt.Errorf("issue session: %w", domain.ErrInvalidConfiguration)
	}
	refreshToken, hash, err := randomToken()
	if err != nil {
		return nil, err
	}
	if familyID == "" {
		familyID, err = randomID()
		if err != nil {
			return nil, fmt.Errorf("generate token family: %w", err)
		}
	}
	tokenID, err := randomID()
	if err != nil {
		return nil, fmt.Errorf("generate token id: %w", err)
	}
	now := time.Now().UTC()
	if err := s.tokenRepository.Create(ctx, &domain.AuthToken{
		ID: tokenID, UserID: user.ID.Hex(), TokenHash: hash, FamilyID: familyID,
		Type: domain.AuthTokenRefresh, Device: device, IPAddress: ipAddress,
		CreatedAt: now, ExpiresAt: now.Add(s.refreshTokenDuration),
	}); err != nil {
		return nil, err
	}
	return &AuthResult{
		User: user, AccessToken: accessToken, RefreshToken: refreshToken,
		ExpiresIn:        int64(s.jwtManager.AccessTokenDuration().Seconds()),
		RefreshExpiresIn: int64(s.refreshTokenDuration.Seconds()),
	}, nil
}

func ensureAccountActive(user *domain.User) error {
	if user == nil {
		return domain.ErrInvalidCredentials
	}
	switch user.Status {
	case "", domain.UserStatusActive:
		return nil
	case domain.UserStatusPendingVerification:
		return domain.ErrAccountPending
	case domain.UserStatusSuspended:
		return domain.ErrAccountSuspended
	default:
		return domain.ErrAccountDeleted
	}
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

	status := domain.UserStatusActive
	if s.requireEmailVerification {
		status = domain.UserStatusPendingVerification
	}
	user := &domain.User{
		ID:           primitive.NewObjectID(),
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         domain.UserRoleUser,
		Status:       status,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	if s.requireEmailVerification {
		if err := s.sendActionToken(ctx, user, domain.AuthTokenEmailVerify, 24*time.Hour); err != nil {
			return nil, err
		}
		return &AuthResult{User: user}, nil
	}
	return s.issueSession(ctx, user, "", "", "")
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	if email == "" || input.Password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}

		return nil, err
	}

	if user.PasswordHash == "" {
		return nil, domain.ErrInvalidCredentials
	}

	if err := ensureAccountActive(user); err != nil {
		return nil, err
	}

	if err := comparePassword(user.PasswordHash, input.Password); err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	return s.issueSession(ctx, user, "", input.Device, input.IPAddress)
}

func (s *AuthService) Refresh(ctx context.Context, input RefreshInput) (*AuthResult, error) {
	if strings.TrimSpace(input.RefreshToken) == "" || s.tokenRepository == nil {
		return nil, domain.ErrInvalidToken
	}
	stored, err := s.tokenRepository.FindByHash(ctx, tokenHash(input.RefreshToken), domain.AuthTokenRefresh)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	now := time.Now().UTC()
	if stored.RevokedAt != nil {
		_ = s.tokenRepository.RevokeFamily(ctx, stored.FamilyID, "refresh token reuse detected", now)
		return nil, domain.ErrTokenRevoked
	}
	if !now.Before(stored.ExpiresAt) {
		return nil, domain.ErrExpiredToken
	}
	if err := s.tokenRepository.Revoke(ctx, stored.ID, "rotated", now); err != nil {
		return nil, domain.ErrInvalidToken
	}
	user, err := s.userRepository.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	return s.issueSession(ctx, user, stored.FamilyID, input.Device, input.IPAddress)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string, allSessions bool) error {
	if strings.TrimSpace(refreshToken) == "" || s.tokenRepository == nil {
		return domain.ErrInvalidToken
	}
	stored, err := s.tokenRepository.FindByHash(ctx, tokenHash(refreshToken), domain.AuthTokenRefresh)
	if err != nil {
		return domain.ErrInvalidToken
	}
	now := time.Now().UTC()
	if allSessions {
		return s.tokenRepository.RevokeUser(ctx, stored.UserID, "logout all sessions", now)
	}
	if stored.RevokedAt != nil {
		return nil
	}
	return s.tokenRepository.Revoke(ctx, stored.ID, "logout", now)
}

func (s *AuthService) sendActionToken(ctx context.Context, user *domain.User, tokenType domain.AuthTokenType, duration time.Duration) error {
	if s.tokenRepository == nil || s.notifier == nil {
		return fmt.Errorf("send action token: %w", domain.ErrInvalidConfiguration)
	}
	raw, hash, err := randomToken()
	if err != nil {
		return err
	}
	id, err := randomID()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if err := s.tokenRepository.Create(ctx, &domain.AuthToken{
		ID: id, UserID: user.ID.Hex(), TokenHash: hash, Type: tokenType,
		CreatedAt: now, ExpiresAt: now.Add(duration),
	}); err != nil {
		return err
	}
	if tokenType == domain.AuthTokenEmailVerify {
		return s.notifier.SendEmailVerification(ctx, user.Email, raw)
	}
	return s.notifier.SendPasswordReset(ctx, user.Email, raw)
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepository.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, domain.ErrUserNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.sendActionToken(ctx, user, domain.AuthTokenPasswordReset, time.Hour)
}

func (s *AuthService) RequestEmailVerification(ctx context.Context, email string) error {
	user, err := s.userRepository.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, domain.ErrUserNotFound) {
		return nil
	}
	if err != nil || user.Status != domain.UserStatusPendingVerification {
		return err
	}
	return s.sendActionToken(ctx, user, domain.AuthTokenEmailVerify, 24*time.Hour)
}

func (s *AuthService) consumeActionToken(ctx context.Context, raw string, tokenType domain.AuthTokenType) (*domain.AuthToken, *domain.User, error) {
	stored, err := s.tokenRepository.FindByHash(ctx, tokenHash(strings.TrimSpace(raw)), tokenType)
	if err != nil || !stored.Active(time.Now().UTC()) {
		return nil, nil, domain.ErrInvalidToken
	}
	user, err := s.userRepository.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, nil, domain.ErrInvalidToken
	}
	return stored, user, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, rawToken string) error {
	stored, user, err := s.consumeActionToken(ctx, rawToken, domain.AuthTokenEmailVerify)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	user.Status = domain.UserStatusActive
	user.EmailVerifiedAt = &now
	user.UpdatedAt = now
	if err := s.userRepository.Update(ctx, user); err != nil {
		return err
	}
	return s.tokenRepository.Revoke(ctx, stored.ID, "email verified", now)
}

func (s *AuthService) ResetPassword(ctx context.Context, rawToken, password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}
	stored, user, err := s.consumeActionToken(ctx, rawToken, domain.AuthTokenPasswordReset)
	if err != nil {
		return err
	}
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	user.PasswordHash = hash
	user.UpdatedAt = now
	if err := s.userRepository.Update(ctx, user); err != nil {
		return err
	}
	if err := s.tokenRepository.Revoke(ctx, stored.ID, "password reset", now); err != nil {
		return err
	}
	return s.tokenRepository.RevokeUser(ctx, user.ID.Hex(), "password changed", now)
}
