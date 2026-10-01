package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	userv1 "github.com/Ishan-Gijavanekar/user-service/api/proto"
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	transportgrpc "github.com/Ishan-Gijavanekar/user-service/internal/transport/grpc"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeUserRepository struct {
	createErr error
	getErr    error
	listErr   error
	updateErr error
	deleteErr error

	createdUser *domain.User
	updatedUser *domain.User
	deletedID   string
	listParams  repository.ListUserParmas

	user  *domain.User
	users []*domain.User
	total int64
}

func (f *fakeUserRepository) Create(_ context.Context, user *domain.User) error {
	f.createdUser = user
	return f.createErr
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

func newUserHandler(repo *fakeUserRepository) *transportgrpc.UserHandler {
	var userRepository repository.UserRespository = repo
	userService := service.NewUserService(&userRepository)
	return transportgrpc.NewUserHandler(userService)
}

func testUser() *domain.User {
	now := time.Now().UTC()
	return &domain.User{
		ID:        primitive.NewObjectID(),
		Name:      "Ada Lovelace",
		Email:     "ada@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestCreateUser(t *testing.T) {
	repo := &fakeUserRepository{}
	handler := newUserHandler(repo)

	resp, err := handler.CreateUser(context.Background(), &userv1.CreateUserRequest{
		Name:  " Ada Lovelace ",
		Email: " ADA@EXAMPLE.COM ",
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	if resp.GetUser().GetName() != "Ada Lovelace" {
		t.Fatalf("expected trimmed name, got %q", resp.GetUser().GetName())
	}
	if resp.GetUser().GetEmail() != "ada@example.com" {
		t.Fatalf("expected normalized request email, got %q", resp.GetUser().GetEmail())
	}
	if repo.createdUser == nil {
		t.Fatal("expected repository Create to be called")
	}
}

func TestCreateUserValidationAndErrors(t *testing.T) {
	tests := []struct {
		name string
		req  *userv1.CreateUserRequest
		err  error
		code codes.Code
	}{
		{name: "nil request", req: nil, code: codes.InvalidArgument},
		{name: "missing name", req: &userv1.CreateUserRequest{Email: "ada@example.com"}, code: codes.InvalidArgument},
		{name: "missing email", req: &userv1.CreateUserRequest{Name: "Ada"}, code: codes.InvalidArgument},
		{
			name: "duplicate email",
			req:  &userv1.CreateUserRequest{Name: "Ada", Email: "ada@example.com"},
			err:  domain.ErrUserAlreadyExsists,
			code: codes.AlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newUserHandler(&fakeUserRepository{createErr: tt.err})
			_, err := handler.CreateUser(context.Background(), tt.req)
			if status.Code(err) != tt.code {
				t.Fatalf("expected %s, got %s (%v)", tt.code, status.Code(err), err)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	user := testUser()
	handler := newUserHandler(&fakeUserRepository{user: user})

	resp, err := handler.GetUser(context.Background(), &userv1.GetUserRequest{Id: user.ID.Hex()})
	if err != nil {
		t.Fatalf("GetUser returned error: %v", err)
	}

	if resp.GetUser().GetId() != user.ID.Hex() {
		t.Fatalf("expected id %q, got %q", user.ID.Hex(), resp.GetUser().GetId())
	}
}

func TestGetUserValidationAndErrors(t *testing.T) {
	tests := []struct {
		name string
		req  *userv1.GetUserRequest
		err  error
		code codes.Code
	}{
		{name: "nil request", req: nil, code: codes.InvalidArgument},
		{name: "missing id", req: &userv1.GetUserRequest{}, code: codes.InvalidArgument},
		{name: "not found", req: &userv1.GetUserRequest{Id: primitive.NewObjectID().Hex()}, err: domain.ErrUserNotFound, code: codes.NotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newUserHandler(&fakeUserRepository{getErr: tt.err})
			_, err := handler.GetUser(context.Background(), tt.req)
			if status.Code(err) != tt.code {
				t.Fatalf("expected %s, got %s (%v)", tt.code, status.Code(err), err)
			}
		})
	}
}

func TestGetUsers(t *testing.T) {
	users := []*domain.User{testUser(), testUser()}
	repo := &fakeUserRepository{
		users: users,
		total: 7,
	}
	handler := newUserHandler(repo)

	resp, err := handler.GetUsers(context.Background(), &userv1.ListUserRequest{Page: 2, Limit: 5})
	if err != nil {
		t.Fatalf("GetUsers returned error: %v", err)
	}

	if len(resp.GetUsers()) != len(users) {
		t.Fatalf("expected %d users, got %d", len(users), len(resp.GetUsers()))
	}
	if resp.GetTotal() != 7 {
		t.Fatalf("expected total from service result, got %d", resp.GetTotal())
	}
	if resp.GetPage() != 2 || resp.GetLimit() != 5 {
		t.Fatalf("expected page/limit 2/5, got %d/%d", resp.GetPage(), resp.GetLimit())
	}
}

func TestGetUsersValidationAndErrors(t *testing.T) {
	handler := newUserHandler(&fakeUserRepository{})
	if _, err := handler.GetUsers(context.Background(), nil); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for nil request, got %s (%v)", status.Code(err), err)
	}

	handler = newUserHandler(&fakeUserRepository{listErr: errors.New("database failed")})
	if _, err := handler.GetUsers(context.Background(), &userv1.ListUserRequest{}); status.Code(err) != codes.Unknown {
		t.Fatalf("expected Unknown for repository error, got %s (%v)", status.Code(err), err)
	}
}

func TestUpdateUser(t *testing.T) {
	user := testUser()
	repo := &fakeUserRepository{user: user}
	handler := newUserHandler(repo)

	_, err := handler.UpdateUser(context.Background(), &userv1.UpdateUserRequest{
		Id:    user.ID.Hex(),
		Name:  " Grace Hopper ",
		Email: " GRACE@EXAMPLE.COM ",
	})
	if err != nil {
		t.Fatalf("UpdateUser returned error: %v", err)
	}

	if repo.updatedUser == nil {
		t.Fatal("expected repository Update to be called")
	}
	if repo.updatedUser.ID != user.ID {
		t.Fatalf("expected updated user id %s, got %s", user.ID.Hex(), repo.updatedUser.ID.Hex())
	}
	if repo.updatedUser.Name != "Grace Hopper" || repo.updatedUser.Email != "grace@example.com" {
		t.Fatalf("unexpected updated user: %#v", repo.updatedUser)
	}
}

func TestUpdateUserValidationAndErrors(t *testing.T) {
	tests := []struct {
		name string
		req  *userv1.UpdateUserRequest
		err  error
		code codes.Code
	}{
		{name: "nil request", req: nil, code: codes.InvalidArgument},
		{name: "missing fields", req: &userv1.UpdateUserRequest{Id: primitive.NewObjectID().Hex(), Name: "Ada"}, code: codes.InvalidArgument},
		{name: "not found", req: &userv1.UpdateUserRequest{Id: primitive.NewObjectID().Hex(), Name: "Ada", Email: "ada@example.com"}, err: domain.ErrInvalidUserId, code: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newUserHandler(&fakeUserRepository{getErr: tt.err})
			_, err := handler.UpdateUser(context.Background(), tt.req)
			if status.Code(err) != tt.code {
				t.Fatalf("expected %s, got %s (%v)", tt.code, status.Code(err), err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	id := primitive.NewObjectID().Hex()
	repo := &fakeUserRepository{}
	handler := newUserHandler(repo)

	if _, err := handler.DeleteUser(context.Background(), &userv1.DeleteUserRequest{Id: id}); err != nil {
		t.Fatalf("DeleteUser returned error: %v", err)
	}
	if repo.deletedID != id {
		t.Fatalf("expected deleted id %q, got %q", id, repo.deletedID)
	}
}

func TestDeleteUserValidationAndErrors(t *testing.T) {
	tests := []struct {
		name string
		req  *userv1.DeleteUserRequest
		err  error
		code codes.Code
	}{
		{name: "nil request", req: nil, code: codes.InvalidArgument},
		{name: "missing id", req: &userv1.DeleteUserRequest{}, code: codes.InvalidArgument},
		{name: "invalid id", req: &userv1.DeleteUserRequest{Id: "bad-id"}, err: domain.ErrInvalidUserId, code: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newUserHandler(&fakeUserRepository{deleteErr: tt.err})
			_, err := handler.DeleteUser(context.Background(), tt.req)
			if status.Code(err) != tt.code {
				t.Fatalf("expected %s, got %s (%v)", tt.code, status.Code(err), err)
			}
		})
	}
}
