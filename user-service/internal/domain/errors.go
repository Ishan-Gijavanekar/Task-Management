package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("User not found")
	ErrUserAlreadyExsists = errors.New("User already exsists")
	ErrInvalidUserId      = errors.New("Invalid User Id")
	ErrInvalidName        = errors.New("Invalid name")
	ErrInvalidEmail       = errors.New("Invalid Name")
)
