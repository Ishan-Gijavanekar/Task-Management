package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
	"github.com/Ishan-Gijavanekar/user-service/internal/testutil"
)

type captureAuthNotifier struct {
	verificationToken string
	resetToken        string
}

func (n *captureAuthNotifier) SendEmailVerification(_ context.Context, _ string, token string) error {
	n.verificationToken = token
	return nil
}

func (n *captureAuthNotifier) SendPasswordReset(_ context.Context, _ string, token string) error {
	n.resetToken = token
	return nil
}

func newLifecycleAuthService(requireVerification bool) (*AuthService, *captureAuthNotifier) {
	userRepo := testutil.NewMemoryUserRepository()
	var users repository.UserRespository = userRepo
	notifier := &captureAuthNotifier{}
	return NewAuthService(users, auth.NewJWTManager("secret", "test", 15*time.Minute),
		WithAuthTokenRepository(testutil.NewMemoryAuthTokenRepository()),
		WithRefreshTokenDuration(24*time.Hour),
		WithEmailVerificationRequired(requireVerification),
		WithAuthNotificationSender(notifier),
	), notifier
}

func TestRefreshTokenRotationAndReuseDetection(t *testing.T) {
	service, _ := newLifecycleAuthService(false)
	ctx := context.Background()
	registered, err := service.Register(ctx, RegisterInput{Name: "Ada", Email: "ada@example.com", Password: "Password123"})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}
	if registered.RefreshToken == "" {
		t.Fatal("Register() returned no refresh token")
	}

	rotated, err := service.Refresh(ctx, RefreshInput{RefreshToken: registered.RefreshToken, Device: "test"})
	if err != nil {
		t.Fatalf("Refresh() error: %v", err)
	}
	if rotated.RefreshToken == registered.RefreshToken || rotated.AccessToken == "" {
		t.Fatal("Refresh() did not rotate the session")
	}
	if _, err := service.Refresh(ctx, RefreshInput{RefreshToken: registered.RefreshToken}); !errors.Is(err, domain.ErrTokenRevoked) {
		t.Fatalf("reused Refresh() error = %v, want ErrTokenRevoked", err)
	}
	if _, err := service.Refresh(ctx, RefreshInput{RefreshToken: rotated.RefreshToken}); !errors.Is(err, domain.ErrTokenRevoked) {
		t.Fatalf("family Refresh() error = %v, want ErrTokenRevoked", err)
	}
}

func TestEmailVerificationAndPasswordReset(t *testing.T) {
	service, notifier := newLifecycleAuthService(true)
	ctx := context.Background()
	registered, err := service.Register(ctx, RegisterInput{Name: "Ada", Email: "ada@example.com", Password: "Password123"})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}
	if registered.AccessToken != "" || notifier.verificationToken == "" {
		t.Fatal("pending registration did not require email verification")
	}
	if _, err := service.Login(ctx, LoginInput{Email: "ada@example.com", Password: "Password123"}); !errors.Is(err, domain.ErrAccountPending) {
		t.Fatalf("Login() error = %v, want ErrAccountPending", err)
	}
	if err := service.VerifyEmail(ctx, notifier.verificationToken); err != nil {
		t.Fatalf("VerifyEmail() error: %v", err)
	}
	if _, err := service.Login(ctx, LoginInput{Email: "ada@example.com", Password: "Password123"}); err != nil {
		t.Fatalf("Login() after verification error: %v", err)
	}

	if err := service.RequestPasswordReset(ctx, "ada@example.com"); err != nil {
		t.Fatalf("RequestPasswordReset() error: %v", err)
	}
	if err := service.ResetPassword(ctx, notifier.resetToken, "NewPassword123"); err != nil {
		t.Fatalf("ResetPassword() error: %v", err)
	}
	if _, err := service.Login(ctx, LoginInput{Email: "ada@example.com", Password: "Password123"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("old password Login() error = %v", err)
	}
	if _, err := service.Login(ctx, LoginInput{Email: "ada@example.com", Password: "NewPassword123"}); err != nil {
		t.Fatalf("new password Login() error: %v", err)
	}
}
