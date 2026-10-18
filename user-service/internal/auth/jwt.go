package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	activeKeyID         string
	keys                map[string][]byte
	issuer              string
	accessTokenDuration time.Duration
}

func NewJWTManager(secret, issuer string, accessTokenDuration time.Duration) *JWTManager {
	return NewRotatingJWTManager("current", map[string]string{"current": secret}, issuer, accessTokenDuration)
}

func NewRotatingJWTManager(activeKeyID string, keys map[string]string, issuer string, accessTokenDuration time.Duration) *JWTManager {
	parsedKeys := make(map[string][]byte, len(keys))
	for keyID, secret := range keys {
		if keyID != "" && secret != "" {
			parsedKeys[keyID] = []byte(secret)
		}
	}
	return &JWTManager{
		activeKeyID:         activeKeyID,
		keys:                parsedKeys,
		issuer:              issuer,
		accessTokenDuration: accessTokenDuration,
	}
}

func (m *JWTManager) Validate() error {
	if m == nil || m.issuer == "" {
		return fmt.Errorf("jwt manager: %w", domain.ErrInvalidConfiguration)
	}
	if m.activeKeyID == "" || len(m.keys[m.activeKeyID]) == 0 {
		return fmt.Errorf("jwt manager: %w: active signing key is missing", domain.ErrInvalidConfiguration)
	}
	return nil
}

func (m *JWTManager) GenerateAccessToken(user *domain.User) (string, error) {
	if user == nil {
		return "", fmt.Errorf("failed to generate token user is nil")
	}
	if err := m.Validate(); err != nil {
		return "", err
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
	token.Header["kid"] = m.activeKeyID
	signedToken, err := token.SignedString(m.keys[m.activeKeyID])
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

			keyID, _ := token.Header["kid"].(string)
			if keyID == "" {
				keyID = m.activeKeyID
			}
			key, ok := m.keys[keyID]
			if !ok {
				return nil, domain.ErrInvalidToken
			}
			return key, nil
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
