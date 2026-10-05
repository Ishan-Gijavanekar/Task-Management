package interceptor

import (
	"context"
	"testing"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthInterceptorAllowsPublicMethod(t *testing.T) {
	interceptor := NewAuthInterceptor(nil, []string{"/auth.v1.AuthService/Login"})

	called := false
	_, err := interceptor.Unary()(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/Login"},
		func(ctx context.Context, req any) (any, error) {
			called = true
			return "ok", nil
		},
	)
	if err != nil {
		t.Fatalf("Unary() error = %v", err)
	}
	if !called {
		t.Fatal("expected public method handler to be called")
	}
}

func TestAuthInterceptorRequiresAuthorizationHeader(t *testing.T) {
	interceptor := NewAuthInterceptor(testJWTManager(), nil)

	_, err := interceptor.Unary()(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/user.v1.UserService/GetUser"},
		func(ctx context.Context, req any) (any, error) {
			t.Fatal("handler should not be called")
			return nil, nil
		},
	)
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %s (%v)", status.Code(err), err)
	}
}

func TestAuthInterceptorRejectsMalformedAuthorizationHeader(t *testing.T) {
	tests := []string{
		"",
		"Bearer",
		"Basic token",
		"Bearer one two",
	}

	for _, header := range tests {
		t.Run(header, func(t *testing.T) {
			interceptor := NewAuthInterceptor(testJWTManager(), nil)
			ctx := metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs(authorizationMetadataKey, header),
			)

			_, err := interceptor.Unary()(
				ctx,
				nil,
				&grpc.UnaryServerInfo{FullMethod: "/user.v1.UserService/GetUser"},
				func(ctx context.Context, req any) (any, error) {
					t.Fatal("handler should not be called")
					return nil, nil
				},
			)
			if status.Code(err) != codes.Unauthenticated {
				t.Fatalf("expected Unauthenticated, got %s (%v)", status.Code(err), err)
			}
		})
	}
}

func TestAuthInterceptorRejectsInvalidToken(t *testing.T) {
	interceptor := NewAuthInterceptor(testJWTManager(), nil)
	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(authorizationMetadataKey, "Bearer invalid-token"),
	)

	_, err := interceptor.Unary()(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/user.v1.UserService/GetUser"},
		func(ctx context.Context, req any) (any, error) {
			t.Fatal("handler should not be called")
			return nil, nil
		},
	)
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %s (%v)", status.Code(err), err)
	}
}

func TestAuthInterceptorAddsAuthenticatedUserToContext(t *testing.T) {
	manager := testJWTManager()
	user := &domain.User{
		ID:    primitive.NewObjectID(),
		Email: "ada@example.com",
		Role:  domain.UserRoleAdmin,
	}
	token, err := manager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	interceptor := NewAuthInterceptor(manager, nil)
	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(authorizationMetadataKey, "bearer "+token),
	)

	_, err = interceptor.Unary()(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/user.v1.UserService/GetUser"},
		func(ctx context.Context, req any) (any, error) {
			authenticatedUser, ok := auth.AuthenticatedUserFromContext(ctx)
			if !ok {
				t.Fatal("expected authenticated user in context")
			}
			if authenticatedUser.ID != user.ID.Hex() {
				t.Fatalf("expected user ID %q, got %q", user.ID.Hex(), authenticatedUser.ID)
			}
			if authenticatedUser.Email != user.Email {
				t.Fatalf("expected email %q, got %q", user.Email, authenticatedUser.Email)
			}
			if authenticatedUser.Role != user.Role {
				t.Fatalf("expected role %q, got %q", user.Role, authenticatedUser.Role)
			}
			return "ok", nil
		},
	)
	if err != nil {
		t.Fatalf("Unary() error = %v", err)
	}
}

func TestAuthInterceptorMissingJWTManager(t *testing.T) {
	interceptor := NewAuthInterceptor(nil, nil)

	_, err := interceptor.Unary()(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/user.v1.UserService/GetUser"},
		func(ctx context.Context, req any) (any, error) {
			t.Fatal("handler should not be called")
			return nil, nil
		},
	)
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal, got %s (%v)", status.Code(err), err)
	}
}

func testJWTManager() *auth.JWTManager {
	return auth.NewJWTManager("test-secret", "user-service", 15*time.Minute)
}
