package service

import (
	"net/mail"
	"strings"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
)

func ValidateUser(name, email string) error {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)

	if name == "" || len(name) > 100 {
		return domain.ErrInvalidName
	}
	if email == "" || len(email) > 256 {
		return domain.ErrInvalidEmail
	}

	address, err := mail.ParseAddress(email)
	if err != nil {
		return domain.ErrInvalidEmail
	}

	if address.Address != email {
		return domain.ErrInvalidEmail
	}

	return nil
}
