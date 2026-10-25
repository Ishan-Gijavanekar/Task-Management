package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateAuthorization(t *testing.T) {
	claims := Claims{UserID: "user-1", Role: "user", RegisteredClaims: jwt.RegisteredClaims{Issuer: "issuer", Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = "current"
	raw, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := New(map[string]string{"current": "secret"}, "issuer").ValidateAuthorization("Bearer " + raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.UserID != "user-1" {
		t.Fatalf("got %s", got.UserID)
	}
}
func TestRejectsMalformedUnknownAndExpiredTokens(t *testing.T) {
	v := New(map[string]string{"current": "secret"}, "issuer")
	for _, header := range []string{"", "Basic abc", "Bearer invalid"} {
		if _, err := v.ValidateAuthorization(header); !errors.Is(err, ErrUnauthenticated) {
			t.Fatalf("%q: %v", header, err)
		}
	}
}
