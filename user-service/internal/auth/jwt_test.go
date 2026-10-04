package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestJWTManagerGenerateAndValidate(
	t *testing.T,
) {

	manager := NewJWTManager(
		"test-secret",
		"user-service",
		15*time.Minute,
	)

	userID := primitive.NewObjectID()

	user := &domain.User{
		ID:    userID,
		Name:  "John Doe",
		Email: "john@example.com",
		Role:  domain.UserRoleUser,
	}

	token, err := manager.GenerateAccessToken(
		user,
	)

	if err != nil {
		t.Fatalf(
			"GenerateAccessToken() error = %v",
			err,
		)
	}

	if token == "" {
		t.Fatal(
			"GenerateAccessToken() returned empty token",
		)
	}

	claims, err := manager.ValidateAccessToken(
		token,
	)

	if err != nil {
		t.Fatalf(
			"ValidateAccessToken() error = %v",
			err,
		)
	}

	if claims.UserID != userID.Hex() {
		t.Errorf(
			"UserID = %s, want %s",
			claims.UserID,
			userID.Hex(),
		)
	}

	if claims.Email != user.Email {
		t.Errorf(
			"Email = %s, want %s",
			claims.Email,
			user.Email,
		)
	}

	if claims.Role != domain.UserRoleUser {
		t.Errorf(
			"Role = %s, want %s",
			claims.Role,
			domain.UserRoleUser,
		)
	}
}

func TestJWTManagerInvalidToken(
	t *testing.T,
) {

	manager := NewJWTManager(
		"test-secret",
		"user-service",
		15*time.Minute,
	)

	_, err := manager.ValidateAccessToken(
		"invalid-token",
	)

	if !errors.Is(
		err,
		domain.ErrInvalidToken,
	) {

		t.Fatalf(
			"expected ErrInvalidToken, got %v",
			err,
		)
	}
}

func TestJWTManagerExpiredToken(
	t *testing.T,
) {

	manager := NewJWTManager(
		"test-secret",
		"user-service",
		-time.Minute,
	)

	user := &domain.User{
		ID:    primitive.NewObjectID(),
		Name:  "John Doe",
		Email: "john@example.com",
		Role:  domain.UserRoleUser,
	}

	token, err := manager.GenerateAccessToken(
		user,
	)

	if err != nil {
		t.Fatalf(
			"GenerateAccessToken() error = %v",
			err,
		)
	}

	_, err = manager.ValidateAccessToken(
		token,
	)

	if !errors.Is(
		err,
		domain.ErrExpiredToken,
	) {

		t.Fatalf(
			"expected ErrExpiredToken, got %v",
			err,
		)
	}
}
