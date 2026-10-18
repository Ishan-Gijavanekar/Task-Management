package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExsists = errors.New("user already exists")
	ErrInvalidUserId      = errors.New("invalid user ID")
	ErrInvalidName        = errors.New("invalid name")
	ErrInvalidEmail       = errors.New("invalid email")

	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")

	ErrInvalidToken         = errors.New("invalid token")
	ErrExpiredToken         = errors.New("token expired")
	ErrTokenRevoked         = errors.New("token revoked")
	ErrAccountPending       = errors.New("account verification required")
	ErrAccountSuspended     = errors.New("account suspended")
	ErrAccountDeleted       = errors.New("account deleted")
	ErrVersionConflict      = errors.New("resource version conflict")
	ErrInvalidConfiguration = errors.New("invalid configuration")
)
