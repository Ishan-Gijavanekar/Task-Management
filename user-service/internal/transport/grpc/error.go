package grpc

import (
	"errors"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidUserId):
		return status.Error(codes.NotFound, "User Not Found")
	case errors.Is(err, domain.ErrUserAlreadyExsists):
		return status.Error(codes.AlreadyExists, "User Already exsists")
	case errors.Is(err, domain.ErrInvalidUserId):
		return status.Error(codes.InvalidArgument, "User Id does not exsist")
	default:
		return status.Error(codes.Unknown, "Internal Server Error")
	}
}
