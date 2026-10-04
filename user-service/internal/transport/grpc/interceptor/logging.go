package interceptor

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func LoggingUnaryInterceptor(logger slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		response, err := handler(ctx, req)

		duration := time.Since(start)

		logger.Info(
			"grpc request completed",
			"method",
			info.FullMethod,
			"request-id",
			requestIDFromMetadata(ctx),
			"status",
			status.Code(err).String(),
			"duration",
			duration.String(),
		)

		return response, err
	}
}
