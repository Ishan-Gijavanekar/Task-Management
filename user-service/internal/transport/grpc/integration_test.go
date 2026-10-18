package grpc

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	userv1 "github.com/Ishan-Gijavanekar/user-service/api/proto"
	authv1 "github.com/Ishan-Gijavanekar/user-service/api/proto/auth/v1"
	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	"github.com/Ishan-Gijavanekar/user-service/internal/testutil"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func newBufconnClients(t *testing.T) (authv1.AuthServiceClient, userv1.UserServiceClient) {
	t.Helper()
	repo := testutil.NewMemoryUserRepository()
	tokenRepo := testutil.NewMemoryAuthTokenRepository()
	var userRepo repository.UserRespository = repo
	userService := service.NewUserService(&userRepo)
	jwtManager := auth.NewJWTManager("grpc-test-secret", "user-service-test", 15*time.Minute)
	authService := service.NewAuthService(userRepo, jwtManager, service.WithAuthTokenRepository(tokenRepo))
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewServer("0", false, userService, authService, jwtManager, logger)

	listener := bufconn.Listen(1024 * 1024)
	go func() {
		_ = server.server.Serve(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := googlegrpc.DialContext(ctx, "bufnet",
		googlegrpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		googlegrpc.WithTransportCredentials(insecure.NewCredentials()),
		googlegrpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("dial bufconn: %v", err)
	}
	t.Cleanup(func() {
		conn.Close()
		server.server.Stop()
		listener.Close()
	})
	return authv1.NewAuthServiceClient(conn), userv1.NewUserServiceClient(conn)
}

func TestGRPCAuthenticationAndAuthorization(t *testing.T) {
	authClient, userClient := newBufconnClients(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	registered, err := authClient.Register(ctx, &authv1.RegisterRequest{
		Name: "Ada Lovelace", Email: "ada@example.com", Password: "Password123",
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}
	if registered.GetAccessToken() == "" || registered.GetExpiresIn() != 900 {
		t.Fatalf("unexpected auth response: %#v", registered)
	}

	_, err = userClient.GetUser(ctx, &userv1.GetUserRequest{Id: registered.GetUser().GetId()})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("GetUser() without token code = %v, want Unauthenticated", status.Code(err))
	}

	authenticated := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+registered.GetAccessToken())
	response, err := userClient.GetUser(authenticated, &userv1.GetUserRequest{Id: registered.GetUser().GetId()})
	if err != nil {
		t.Fatalf("GetUser() error: %v", err)
	}
	if response.GetUser().GetEmail() != "ada@example.com" {
		t.Fatalf("email = %q, want ada@example.com", response.GetUser().GetEmail())
	}

	_, err = userClient.GetUsers(authenticated, &userv1.ListUserRequest{Page: 1, Limit: 10})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("GetUsers() code = %v, want PermissionDenied", status.Code(err))
	}
}
