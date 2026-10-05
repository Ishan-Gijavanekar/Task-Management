package interceptor

import (
	"context"
	"strings"

	"github.com/Ishan-Gijavanekar/user-service/internal/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const authorizationMetadataKey = "authorization"

type AuthInterceptor struct {
	jwtManager    *auth.JWTManager
	publicMethods map[string]struct{}
}

func NewAuthInterceptor(jwtManager *auth.JWTManager, publicMethods []string) *AuthInterceptor {
	methods := make(map[string]struct{}, len(publicMethods))
	for _, method := range publicMethods {
		methods[method] = struct{}{}
	}

	return &AuthInterceptor{
		jwtManager:    jwtManager,
		publicMethods: methods,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if i.isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		if i.jwtManager == nil {
			return nil, status.Error(codes.Internal, "authentication is not configured")
		}

		tokenString, err := bearerTokenFromContext(ctx)
		if err != nil {
			return nil, err
		}

		claims, err := i.jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired access token")
		}

		ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{
			ID:    claims.UserID,
			Email: claims.Email,
			Role:  claims.Role,
		})

		return handler(ctx, req)
	}
}

func (i *AuthInterceptor) isPublicMethod(method string) bool {
	_, exists := i.publicMethods[method]
	return exists
}

func bearerTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "authorization metadata is required")
	}

	values := md.Get(authorizationMetadataKey)
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization header is required")
	}

	authorization := strings.TrimSpace(values[0])
	parts := strings.Fields(authorization)
	if len(parts) != 2 {
		return "", status.Error(codes.Unauthenticated, "invalid authorization header")
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", status.Error(codes.Unauthenticated, "authorization scheme must be Bearer")
	}

	if parts[1] == "" {
		return "", status.Error(codes.Unauthenticated, "access token is required")
	}

	return parts[1], nil
}
