package grpc

import (
	"context"
	"strings"

	userv1 "github.com/Ishan-Gijavanekar/user-service/api/proto"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	userService *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{
		userService: service,
	}
}

func (u *UserHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.UserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if strings.TrimSpace(req.GetName()) == "" {
		return nil, status.Error(codes.InvalidArgument, "name cannot be empty")
	}

	if strings.TrimSpace(req.GetEmail()) == "" {
		return nil, status.Error(codes.InvalidArgument, "email cannot be empty")
	}

	user, err := u.userService.Create(ctx, service.CreateUserInput{
		Name:  req.GetName(),
		Email: req.GetName(),
	},
	)
	if err != nil {
		return nil, MapError(err)
	}

	return &userv1.UserResponse{
		User: UserToProto(user),
	}, nil
}

func (u *UserHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.UserResponse, error) {
	if req == nil || strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	user, err := u.userService.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, MapError(err)
	}

	return &userv1.UserResponse{
		User: UserToProto(user),
	}, nil
}

func (u *UserHandler) ListUsers(ctx context.Context, req *userv1.ListUserRequest) (*userv1.ListUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "page and size are required")
	}

	result, err := u.userService.List(ctx, service.ListUserInput{
		Page:  req.GetPage(),
		Limit: req.GetLimit(),
	})
	if err != nil {
		return nil, MapError(err)
	}

	users := make([]*userv1.User, 0, len(result.Users))

	for _, user := range result.Users {
		users = append(users, UserToProto(user))
	}

	return &userv1.ListUserResponse{
		Users: users,
		Total: int64(len(result.Users)),
		Page:  result.Page,
		Limit: result.Limit,
	}, nil
}

func (u *UserHandler) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "req cannot be nil")
	}

	if strings.TrimSpace(req.GetName()) == "" || strings.TrimSpace(req.GetId()) == "" || strings.TrimSpace(req.GetEmail()) == "" {
		return status.Error(codes.InvalidArgument, "ID, name and email are required")
	}

	_, err := u.userService.Update(ctx, req.GetId(), service.UpdateUserInput{
		Name:  req.GetName(),
		Email: req.GetEmail(),
	})
	if err != nil {
		return MapError(err)
	}

	return nil
}

func (u *UserHandler) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "req cannot be nil")
	}

	if strings.TrimSpace(req.GetId()) == "" {
		return status.Error(codes.InvalidArgument, "Id is required")
	}

	err := u.userService.Delete(ctx, req.GetId())
	if err != nil {
		MapError(err)
	}

	return nil
}
