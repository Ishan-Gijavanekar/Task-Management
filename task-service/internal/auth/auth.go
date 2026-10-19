package auth

import (
	"context"
	"strings"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}
type Manager struct {
	keys   map[string][]byte
	issuer string
}

func New(keys map[string]string, issuer string) *Manager {
	parsed := map[string][]byte{}
	for k, v := range keys {
		if k != "" && v != "" {
			parsed[k] = []byte(v)
		}
	}
	return &Manager{keys: parsed, issuer: issuer}
}
func (m *Manager) Validate(raw string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, domain.ErrUnauthenticated
		}
		kid, _ := t.Header["kid"].(string)
		key, ok := m.keys[kid]
		if !ok {
			return nil, domain.ErrUnauthenticated
		}
		return key, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid || claims.UserID == "" || claims.Subject != claims.UserID {
		return "", domain.ErrUnauthenticated
	}
	return claims.UserID, nil
}

type key struct{}

func WithActor(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, key{}, id)
}
func Actor(ctx context.Context) (string, error) {
	id, _ := ctx.Value(key{}).(string)
	if id == "" {
		return "", domain.ErrUnauthenticated
	}
	return id, nil
}
func bearer(v string) (string, error) {
	p := strings.Fields(v)
	if len(p) != 2 || !strings.EqualFold(p[0], "bearer") {
		return "", domain.ErrUnauthenticated
	}
	return p[1], nil
}
func (m *Manager) Fiber() fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw, e := bearer(c.Get(fiber.HeaderAuthorization))
		if e != nil {
			return c.Status(401).JSON(fiber.Map{"error": fiber.Map{"code": "unauthenticated", "message": "valid bearer token required", "requestId": c.GetRespHeader("X-Request-ID")}})
		}
		id, e := m.Validate(raw)
		if e != nil {
			return c.Status(401).JSON(fiber.Map{"error": fiber.Map{"code": "unauthenticated", "message": "valid bearer token required"}})
		}
		c.Locals("actorId", id)
		return c.Next()
	}
}
func (m *Manager) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization required")
		}
		raw, e := bearer(values[0])
		if e != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization")
		}
		id, e := m.Validate(raw)
		if e != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid access token")
		}
		return next(WithActor(ctx, id), req)
	}
}
