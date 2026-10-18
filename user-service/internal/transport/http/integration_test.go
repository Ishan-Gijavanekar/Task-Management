package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/Ishan-Gijavanekar/user-service/internal/config"
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	"github.com/Ishan-Gijavanekar/user-service/internal/testutil"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type httpTestStack struct {
	server *Server
	repo   *testutil.MemoryUserRepository
	jwt    *auth.JWTManager
}

const integrationTestTimeoutMilliseconds = 30_000

func newHTTPTestStack() *httpTestStack {
	repo := testutil.NewMemoryUserRepository()
	var userRepo repository.UserRespository = repo
	userService := service.NewUserService(&userRepo)
	jwtManager := auth.NewJWTManager("integration-test-secret", "user-service-test", 15*time.Minute)
	authService := service.NewAuthService(userRepo, jwtManager)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewServer(config.HTTPConfig{
		Port: "0", AllowedOrigins: "http://localhost:5173", BodyLimitMB: 1,
	}, userService, authService, jwtManager, logger)
	return &httpTestStack{server: server, repo: repo, jwt: jwtManager}
}

func performJSONRequest(t *testing.T, server *Server, method, path string, body any, token string) *nethttp.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reader = bytes.NewReader(encoded)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := server.app.Test(req, integrationTestTimeoutMilliseconds)
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
	return resp
}

func decodeHTTPBody[T any](t *testing.T, resp *nethttp.Response) T {
	t.Helper()
	defer resp.Body.Close()
	var value T
	if err := json.NewDecoder(resp.Body).Decode(&value); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return value
}

func TestHTTPRegisterLoginAndSelfAccess(t *testing.T) {
	stack := newHTTPTestStack()
	register := performJSONRequest(t, stack.server, nethttp.MethodPost, "/api/v1/auth/register", RegisterRequest{
		Name: "Ada Lovelace", Email: "ada@example.com", Password: "Password123",
	}, "")
	if register.StatusCode != fiber.StatusCreated {
		t.Fatalf("register status = %d, want %d", register.StatusCode, fiber.StatusCreated)
	}
	registered := decodeHTTPBody[AuthResponse](t, register)
	if registered.AccessToken == "" || registered.ExpiresIn != 900 {
		t.Fatalf("unexpected registration response: %#v", registered)
	}

	login := performJSONRequest(t, stack.server, nethttp.MethodPost, "/api/v1/auth/login", LoginRequest{
		Email: "ada@example.com", Password: "Password123",
	}, "")
	if login.StatusCode != fiber.StatusOK {
		t.Fatalf("login status = %d, want %d", login.StatusCode, fiber.StatusOK)
	}
	loggedIn := decodeHTTPBody[AuthResponse](t, login)

	get := performJSONRequest(t, stack.server, nethttp.MethodGet, "/api/v1/users/"+registered.User.ID, nil, loggedIn.AccessToken)
	if get.StatusCode != fiber.StatusOK {
		t.Fatalf("get status = %d, want %d", get.StatusCode, fiber.StatusOK)
	}
	user := decodeHTTPBody[UserResponse](t, get)
	if user.Email != "ada@example.com" {
		t.Fatalf("email = %q, want ada@example.com", user.Email)
	}
}

func TestHTTPAuthorization(t *testing.T) {
	stack := newHTTPTestStack()
	unauthenticated := performJSONRequest(t, stack.server, nethttp.MethodGet, "/api/v1/users", nil, "")
	defer unauthenticated.Body.Close()
	if unauthenticated.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d", unauthenticated.StatusCode, fiber.StatusUnauthorized)
	}

	admin := &domain.User{ID: primitive.NewObjectID(), Name: "Admin", Email: "admin@example.com", Role: domain.UserRoleAdmin, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := stack.repo.Create(context.Background(), admin); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	token, err := stack.jwt.GenerateAccessToken(admin)
	if err != nil {
		t.Fatalf("generate admin token: %v", err)
	}
	list := performJSONRequest(t, stack.server, nethttp.MethodGet, "/api/v1/users", nil, token)
	if list.StatusCode != fiber.StatusOK {
		t.Fatalf("admin list status = %d, want %d", list.StatusCode, fiber.StatusOK)
	}
	decodeHTTPBody[ListUserResponse](t, list)
}

func TestHTTPCORSAndRateLimit(t *testing.T) {
	stack := newHTTPTestStack()
	preflight := httptest.NewRequest(nethttp.MethodOptions, "/api/v1/auth/login", nil)
	preflight.Header.Set("Origin", "http://localhost:5173")
	preflight.Header.Set("Access-Control-Request-Method", nethttp.MethodPost)
	resp, err := stack.server.app.Test(preflight, integrationTestTimeoutMilliseconds)
	if err != nil {
		t.Fatalf("preflight error: %v", err)
	}
	resp.Body.Close()
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("allow origin = %q", got)
	}

	for requestNumber := 1; requestNumber <= 11; requestNumber++ {
		response := performJSONRequest(t, stack.server, nethttp.MethodPost, "/api/v1/auth/login", LoginRequest{
			Email: "unknown@example.com", Password: "Password123",
		}, "")
		response.Body.Close()
		want := fiber.StatusUnauthorized
		if requestNumber == 11 {
			want = fiber.StatusTooManyRequests
		}
		if response.StatusCode != want {
			t.Fatalf("request %d status = %d, want %d", requestNumber, response.StatusCode, want)
		}
	}
}
