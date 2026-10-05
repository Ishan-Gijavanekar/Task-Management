package grpc

import (
	"context"
	"strings"

	authv1 "github.com/Ishan-Gijavanekar/user-service/api/proto/auth/v1"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "Request is required")
	}

	if strings.TrimSpace(req.GetName()) == "" {
		return nil, status.Error(codes.InvalidArgument, "Name is required")
	}
	if strings.TrimSpace(req.GetEmail()) == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}
	if strings.TrimSpace(req.GetPassword()) == "" {
		return nil, status.Error(codes.InvalidArgument, "Password is required")
	}

	result, err := h.authService.Register(
		ctx,
		service.RegisterInput{
			Name:     req.GetName(),
			Email:    req.GetEmail(),
			Password: req.GetPassword(),
		},
	)
	if err != nil {
		return nil, MapError(err)
	}

	return &authv1.AuthResponse{
		User:        userToAuthProto(result.User),
		AccessToken: result.AccessToken,
		ExpiresIn:   result.ExpiresIn,
		TokenType:   "Bearer",
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "Request is required")
	}

	if strings.TrimSpace(req.GetEmail()) == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}
	if strings.TrimSpace(req.GetPassword()) == "" {
		return nil, status.Error(codes.InvalidArgument, "Password is required")
	}

	result, err := h.authService.Login(
		ctx,
		service.LoginInput{
			Email:    req.GetEmail(),
			Password: req.GetPassword(),
		},
	)
	if err != nil {
		return nil, MapError(err)
	}

	return &authv1.AuthResponse{
		User:        userToAuthProto(result.User),
		AccessToken: result.AccessToken,
		ExpiresIn:   result.ExpiresIn,
		TokenType:   "Bearer",
	}, nil
}
