package auth

import (
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var ErrUnauthenticated = errors.New("authentication required")

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
type Validator struct {
	keys   map[string][]byte
	issuer string
}

func New(keys map[string]string, issuer string) *Validator {
	parsed := map[string][]byte{}
	for id, key := range keys {
		if id != "" && key != "" {
			parsed[id] = []byte(key)
		}
	}
	return &Validator{keys: parsed, issuer: issuer}
}

func (v *Validator) ValidateAuthorization(header string) (*Claims, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, ErrUnauthenticated
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrUnauthenticated
		}
		kid, _ := token.Header["kid"].(string)
		key, ok := v.keys[kid]
		if !ok {
			return nil, ErrUnauthenticated
		}
		return key, nil
	}, jwt.WithIssuer(v.issuer), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid || claims.UserID == "" || claims.Subject != claims.UserID || claims.Role == "" {
		return nil, ErrUnauthenticated
	}
	return claims, nil
}
