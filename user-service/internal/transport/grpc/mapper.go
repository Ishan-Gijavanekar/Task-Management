package grpc

import (
	userv1 "github.com/Ishan-Gijavanekar/user-service/api/proto"
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func UserToProto(user *domain.User) *userv1.User {
	if user == nil {
		return nil
	}

	return &userv1.User{
		Id:    user.ID.Hex(),
		Name:  user.Name,
		Email: user.Email,
		CreatedAt: timestamppb.New(
			user.CreatedAt,
		),
		UpdatedAt: timestamppb.New(
			user.UpdatedAt,
		),
	}
}
