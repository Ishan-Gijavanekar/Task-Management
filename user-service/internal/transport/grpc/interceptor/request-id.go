package interceptor

import (
	"context"
	"crypto/rand"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const RequestIDMetadataKey = "x-request-id"

type contextKey string

const requestIDContextKey contextKey = "request-id"

func RequestIDUnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	requestID := requestIDFromMetadata(ctx)

	if requestID == "" {
		requestID = newRequestID()
	}

	ctx = context.WithValue(ctx, requestIDContextKey, requestID)
	_ = grpc.SetHeader(ctx, metadata.Pairs(RequestIDMetadataKey, requestID))

	return handler(ctx, req)
}

func requestIDFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get(RequestIDMetadataKey)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ""
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
