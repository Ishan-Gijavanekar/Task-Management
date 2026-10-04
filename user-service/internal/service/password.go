package service

import (
	"errors"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func hashPassword(password string) (string, error) {
	if err := validatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func comparePassword(passwordHash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return domain.ErrInvalidCredentials
		}
		return err
	}

	return nil
}
