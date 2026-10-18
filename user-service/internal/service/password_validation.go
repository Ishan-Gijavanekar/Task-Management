package service

import (
	"strings"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72
)

func validatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return domain.ErrInvalidPassword
	}

	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return domain.ErrInvalidPassword
	}

	return nil
}
