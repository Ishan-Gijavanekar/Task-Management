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
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, domain.ErrUserAlreadyExsists):
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, domain.ErrInvalidUserId):
		return status.Error(codes.InvalidArgument, "invalid user ID")
	case errors.Is(err, domain.ErrInvalidName):
		return status.Error(codes.InvalidArgument, "invalid name")
	case errors.Is(err, domain.ErrInvalidEmail):
		return status.Error(codes.InvalidArgument, "invalid email")
	case errors.Is(err, domain.ErrInvalidPassword):
		return status.Error(codes.InvalidArgument, "invalid password")
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, domain.ErrUnauthorized):
		return status.Error(codes.Unauthenticated, "authentication required")
	case errors.Is(err, domain.ErrForbidden):
		return status.Error(codes.PermissionDenied, "permission denied")
	case errors.Is(err, domain.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, "invalid access token")
	case errors.Is(err, domain.ErrExpiredToken):
		return status.Error(codes.Unauthenticated, "access token expired")
	case errors.Is(err, domain.ErrTokenRevoked):
		return status.Error(codes.Unauthenticated, "token revoked")
	case errors.Is(err, domain.ErrAccountPending):
		return status.Error(codes.PermissionDenied, "email verification required")
	case errors.Is(err, domain.ErrAccountSuspended), errors.Is(err, domain.ErrAccountDeleted):
		return status.Error(codes.PermissionDenied, "account unavailable")
	case errors.Is(err, domain.ErrVersionConflict):
		return status.Error(codes.Aborted, "resource version conflict")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
