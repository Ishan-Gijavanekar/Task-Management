package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type fakeUserRepository struct {
	getErr    error
	listErr   error
	updateErr error
	deleteErr error

	updatedUser *domain.User
	deletedID   string
	listParams  repository.ListUserParmas

	user  *domain.User
	users []*domain.User
	total int64
}

func (f *fakeUserRepository) Create(_ context.Context, _ *domain.User) error {
	return nil
}

func (f *fakeUserRepository) GetByID(_ context.Context, _ string) (*domain.User, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.user, nil
}

func (f *fakeUserRepository) GetByEmail(_ context.Context, _ string) (*domain.User, error) {
	return f.user, nil
}

func (f *fakeUserRepository) List(_ context.Context, params repository.ListUserParmas) ([]*domain.User, int64, error) {
	f.listParams = params
	return f.users, f.total, f.listErr
}

func (f *fakeUserRepository) Update(_ context.Context, user *domain.User) error {
	f.updatedUser = user
	return f.updateErr
}

func (f *fakeUserRepository) Delete(_ context.Context, id string) error {
	f.deletedID = id
	return f.deleteErr
}

func newTestUserHandler(repo *fakeUserRepository) *UserHandler {
	var userRepository repository.UserRespository = repo
	userService := service.NewUserService(&userRepository)
	return NewUserHandler(userService)
}

func newTestUser(role domain.UserRole) *domain.User {
	now := time.Date(2026, 10, 12, 10, 30, 0, 0, time.UTC)
	return &domain.User{
		ID:        primitive.NewObjectID(),
		Name:      "Ada Lovelace",
		Email:     "ada@example.com",
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now.Add(time.Hour),
	}
}

func authenticatedHTTPContext(id string, role domain.UserRole) context.Context {
	return auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
		ID:    id,
		Email: "user@example.com",
		Role:  role,
	})
}

func testApp(ctx context.Context, handler *UserHandler) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.SetUserContext(ctx)
		return c.Next()
	})
	app.Get("/users", handler.ListUsers)
	app.Get("/users/:id", handler.GetUser)
	app.Put("/users/:id", handler.UpdateUsers)
	app.Delete("/users/:id", handler.DeleteUser)
	return app
}

func doRequest(t *testing.T, app *fiber.App, method, path string, body []byte) *nethttp.Response {
	t.Helper()

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := nethttp.NewRequest(method, path, reader)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	return resp
}

func decodeJSON[T any](t *testing.T, resp *nethttp.Response) T {
	t.Helper()
	defer resp.Body.Close()

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	return result
}

func TestUserToHTTPResponseIncludesRole(t *testing.T) {
	user := newTestUser(domain.UserRoleAdmin)

	resp := userToHttpResponse(user)

	if resp.Role != domain.UserRoleAdmin {
		t.Fatalf("expected role %q, got %q", domain.UserRoleAdmin, resp.Role)
	}
	if resp.ID != user.ID.Hex() {
		t.Fatalf("expected id %q, got %q", user.ID.Hex(), resp.ID)
	}
	if resp.CreatedAt != user.CreatedAt.Format(time.RFC3339) {
		t.Fatalf("expected createdAt %q, got %q", user.CreatedAt.Format(time.RFC3339), resp.CreatedAt)
	}
}

func TestGetUser(t *testing.T) {
	user := newTestUser(domain.UserRoleUser)
	handler := newTestUserHandler(&fakeUserRepository{user: user})
	app := testApp(authenticatedHTTPContext(user.ID.Hex(), domain.UserRoleUser), handler)

	resp := doRequest(t, app, nethttp.MethodGet, "/users/"+user.ID.Hex(), nil)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body := decodeJSON[UserResponse](t, resp)
	if body.ID != user.ID.Hex() || body.Role != domain.UserRoleUser {
		t.Fatalf("unexpected response: %#v", body)
	}
}

func TestGetUserRejectsOtherUser(t *testing.T) {
	user := newTestUser(domain.UserRoleUser)
	handler := newTestUserHandler(&fakeUserRepository{user: user})
	app := testApp(authenticatedHTTPContext(primitive.NewObjectID().Hex(), domain.UserRoleUser), handler)

	resp := doRequest(t, app, nethttp.MethodGet, "/users/"+user.ID.Hex(), nil)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected status %d, got %d", fiber.StatusForbidden, resp.StatusCode)
	}
}

func TestListUsers(t *testing.T) {
	users := []*domain.User{
		newTestUser(domain.UserRoleUser),
		newTestUser(domain.UserRoleAdmin),
	}
	repo := &fakeUserRepository{users: users, total: 12}
	handler := newTestUserHandler(repo)
	app := testApp(authenticatedHTTPContext("admin-123", domain.UserRoleAdmin), handler)

	resp := doRequest(t, app, nethttp.MethodGet, "/users?page=2&limit=5", nil)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body := decodeJSON[ListUserResponse](t, resp)
	if len(body.Users) != len(users) {
		t.Fatalf("expected %d users, got %d", len(users), len(body.Users))
	}
	if body.Total != 12 || body.Page != 2 || body.Limit != 5 {
		t.Fatalf("unexpected pagination response: %#v", body)
	}
	if body.Users[1].Role != domain.UserRoleAdmin {
		t.Fatalf("expected second user role %q, got %q", domain.UserRoleAdmin, body.Users[1].Role)
	}
	if repo.listParams.Size != 2 || repo.listParams.Limit != 5 {
		t.Fatalf("expected repository params 2/5, got %d/%d", repo.listParams.Size, repo.listParams.Limit)
	}
}

func TestListUsersRequiresAdmin(t *testing.T) {
	handler := newTestUserHandler(&fakeUserRepository{})
	app := testApp(authenticatedHTTPContext("user-123", domain.UserRoleUser), handler)

	resp := doRequest(t, app, nethttp.MethodGet, "/users", nil)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected status %d, got %d", fiber.StatusForbidden, resp.StatusCode)
	}
}

func TestUpdateUsers(t *testing.T) {
	user := newTestUser(domain.UserRoleUser)
	repo := &fakeUserRepository{user: user}
	handler := newTestUserHandler(repo)
	app := testApp(authenticatedHTTPContext(user.ID.Hex(), domain.UserRoleUser), handler)
	body := []byte(`{"name":" Grace Hopper ","email":" GRACE@EXAMPLE.COM "}`)

	resp := doRequest(t, app, nethttp.MethodPut, "/users/"+user.ID.Hex(), body)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	got := decodeJSON[UserResponse](t, resp)
	if got.Name != "Grace Hopper" || got.Email != "grace@example.com" {
		t.Fatalf("unexpected response: %#v", got)
	}
	if repo.updatedUser == nil {
		t.Fatal("expected repository Update to be called")
	}
}

func TestDeleteUser(t *testing.T) {
	userID := primitive.NewObjectID().Hex()
	repo := &fakeUserRepository{}
	handler := newTestUserHandler(repo)
	app := testApp(authenticatedHTTPContext(userID, domain.UserRoleUser), handler)

	resp := doRequest(t, app, nethttp.MethodDelete, "/users/"+userID, nil)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}
	if repo.deletedID != userID {
		t.Fatalf("expected deleted id %q, got %q", userID, repo.deletedID)
	}
}
