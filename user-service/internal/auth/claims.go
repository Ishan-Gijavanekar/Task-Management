package auth

import (
	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string          `json:"iser_id"`
	Email  string          `json:"email"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}
