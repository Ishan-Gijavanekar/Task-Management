package grpc

import (
	"errors"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, "User Not Found")
	case errors.Is(err, domain.ErrUserAlreadyExsists):
		return status.Error(codes.AlreadyExists, "User Already exsists")
	case errors.Is(err, domain.ErrInvalidUserId):
		return status.Error(codes.InvalidArgument, "User Id does not exsist")
	case errors.Is(err, domain.ErrInvalidName):
		return status.Error(codes.InvalidArgument, "Invalid name")
	case errors.Is(err, domain.ErrInvalidEmail):
		return status.Error(codes.InvalidArgument, "Invalid email")
	case errors.Is(err, domain.ErrInvalidPassword):
		return status.Error(codes.InvalidArgument, "Invalid Password")
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "Invaslid Credentails")
	case errors.Is(err, domain.ErrUnauthorized):
		return status.Error(codes.Unauthenticated, "auththentication required")
	case errors.Is(err, domain.ErrForbidden):
		return status.Error(codes.PermissionDenied, "Permission Denied")
	case errors.Is(err, domain.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, "Invalid access token")
	case errors.Is(err, domain.ErrExpiredToken):
		return status.Error(codes.Unauthenticated, "Access token expired")
	default:
		return status.Error(codes.Unknown, "Internal Server Error")
	}
}
