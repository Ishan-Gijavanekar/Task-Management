package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret              []byte
	issuer              string
	accessTokenDuration time.Duration
}

func NewJWTManager(secret, issuer string, accessTokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secret:              []byte(secret),
		issuer:              issuer,
		accessTokenDuration: accessTokenDuration,
	}
}

func (m *JWTManager) GenerateAccessToken(user *domain.User) (string, error) {
	if user == nil {
		return "", fmt.Errorf("failed to generate token user is nil")
	}

	now := time.Now()
	claims := Claims{
		UserID: user.ID.Hex(),
		Email:  user.Email,
		Role:   user.Role,

		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   user.ID.Hex(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("fail to sign access token %w", err)
	}

	return signedToken, nil
}

func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, domain.ErrInvalidToken
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("Unexpected signing method: %s", token.Method.Alg())
			}

			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrExpiredToken
		}
		return nil, domain.ErrInvalidToken
	}

	if !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	if claims.UserID == "" {
		return nil, domain.ErrInvalidToken
	}

	if claims.Subject == "" {
		return nil, domain.ErrInvalidToken
	}

	if claims.Subject != claims.UserID {
		return nil, domain.ErrInvalidToken
	}

	if !claims.Role.IsValid() {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

func (m *JWTManager) AccessTokenDuration() time.Duration {
	return m.accessTokenDuration
}
