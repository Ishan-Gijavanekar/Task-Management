package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log/slog"
	"sync/atomic"
	"time"
)

var requests atomic.Uint64

func HTTP(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Set("X-Request-ID", id)
		carrier := propagation.HeaderCarrier{}
		for k, v := range c.GetReqHeaders() {
			if len(v) > 0 {
				carrier.Set(k, v[0])
			}
		}
		ctx := otel.GetTextMapPropagator().Extract(c.UserContext(), carrier)
		ctx, span := otel.Tracer("task-service/http").Start(ctx, c.Method()+" "+c.Route().Path)
		defer span.End()
		c.SetUserContext(ctx)
		started := time.Now()
		requests.Add(1)
		err := c.Next()
		logger.Info("http request", "request_id", id, "method", c.Method(), "path", c.Path(), "status", c.Response().StatusCode(), "duration_ms", time.Since(started).Milliseconds())
		return err
	}
}
func Metrics(c *fiber.Ctx) error {
	c.Type("text/plain")
	return c.SendString(fmt.Sprintf("# TYPE task_service_http_requests_total counter\ntask_service_http_requests_total %d\n", requests.Load()))
}
func Unary(logger *slog.Logger, limit int) grpc.UnaryServerInterceptor {
	sem := make(chan struct{}, limit)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		md, _ := metadata.FromIncomingContext(ctx)
		carrier := propagation.HeaderCarrier{}
		for k, v := range md {
			if len(v) > 0 {
				carrier.Set(k, v[0])
			}
		}
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
		ctx, span := otel.Tracer("task-service/grpc").Start(ctx, info.FullMethod)
		defer span.End()
		started := time.Now()
		resp, err := next(ctx, req)
		logger.Info("grpc request", "method", info.FullMethod, "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return resp, err
	}
}
