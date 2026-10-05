package grpc

import (
	authv1 "github.com/Ishan-Gijavanekar/user-service/api/proto/auth/v1"
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func userToAuthProto(user *domain.User) *authv1.AuthUser {
	if user == nil {
		return nil
	}

	return &authv1.AuthUser{
		Id:        user.ID.Hex(),
		Name:      user.Name,
		Email:     user.Email,
		Role:      string(user.Role),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}
