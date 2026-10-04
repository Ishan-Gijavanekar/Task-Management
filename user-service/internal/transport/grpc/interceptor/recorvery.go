package interceptor

import (
	"context"
	"log/slog"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoveryUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error(
					"panic recovered from grpc request",
					"method",
					info.FullMethod,
					"request_id",
					requestIDFromMetadata(ctx),
					"panic",
					recovered,
					"stack",
					string(debug.Stack()),
				)

				err = status.Error(
					codes.Internal,
					"internal server error",
				)
			}
		}()
		return handler(ctx, req)
	}
}
