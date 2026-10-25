package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Ishan-Gijavanekar/gateway-service/internal/auth"
	"github.com/Ishan-Gijavanekar/gateway-service/internal/config"
	"github.com/Ishan-Gijavanekar/gateway-service/internal/gateway"
	"github.com/Ishan-Gijavanekar/gateway-service/internal/ratelimit"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

type Server struct {
	app              *fiber.App
	port             string
	logger           *slog.Logger
	proxy            *gateway.Proxy
	validator        *auth.Validator
	limiter          *ratelimit.Limiter
	readinessTimeout time.Duration
}

var requests atomic.Uint64

func New(cfg *config.Config, proxy *gateway.Proxy, validator *auth.Validator, logger *slog.Logger, openAPI []byte) (*Server, error) {
	if cfg == nil || proxy == nil || validator == nil || logger == nil {
		return nil, fmt.Errorf("configuration, proxy, validator, and logger are required")
	}
	server := &Server{port: cfg.HTTPPort, logger: logger, proxy: proxy, validator: validator, limiter: ratelimit.New(), readinessTimeout: cfg.UpstreamTimeout}
	app := fiber.New(fiber.Config{AppName: "api-gateway", DisableStartupMessage: true, BodyLimit: cfg.BodyLimit, ProxyHeader: fiber.HeaderXForwardedFor, EnableTrustedProxyCheck: len(cfg.TrustedProxies) > 0, TrustedProxies: cfg.TrustedProxies, ErrorHandler: server.errorHandler})
	server.app = app
	app.Use(recover.New())
	app.Use(server.requestContext)
	app.Use(cors.New(cors.Config{AllowOrigins: cfg.PublicOrigin, AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS", AllowHeaders: "Accept,Authorization,Content-Type,Idempotency-Key,X-Request-ID,Traceparent,Tracestate,Baggage", ExposeHeaders: "X-Request-ID,Retry-After,ETag", AllowCredentials: true, MaxAge: 600}))
	app.Get("/livez", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) })
	app.Get("/readyz", server.ready)
	app.Get("/metrics", metrics)
	app.Get("/openapi.json", func(c *fiber.Ctx) error { c.Type("json"); return c.Send(openAPI) })
	app.All("/api/v1/*", server.forward)
	app.Use(func(c *fiber.Ctx) error { return gatewayError(c, 404, "not_found", "route not found", nil) })
	return server, nil
}
func (s *Server) App() *fiber.App                { return s.app }
func (s *Server) Start() error                   { return s.app.Listen(":" + s.port) }
func (s *Server) Stop(ctx context.Context) error { return s.app.ShutdownWithContext(ctx) }

func (s *Server) forward(c *fiber.Ctx) error {
	route, ok := gateway.Match(c.Method(), c.Path())
	if !ok {
		return gatewayError(c, 404, "not_found", "route not found", nil)
	}
	if len(c.Body()) > 0 && !strings.HasPrefix(strings.ToLower(c.Get(fiber.HeaderContentType)), fiber.MIMEApplicationJSON) {
		return gatewayError(c, 415, "unsupported_media_type", "request body must use application/json", nil)
	}
	actorID := ""
	if route.Protected {
		claims, err := s.validator.ValidateAuthorization(c.Get(fiber.HeaderAuthorization))
		if err != nil {
			return gatewayError(c, 401, "unauthenticated", "valid bearer token required", nil)
		}
		actorID = claims.UserID
	}
	key := c.IP() + ":" + route.Name
	if allowed, retry := s.limiter.Allow(key, route.Limit, route.Window); !allowed {
		seconds := int(retry.Seconds())
		if seconds < 1 {
			seconds = 1
		}
		c.Set("Retry-After", strconv.Itoa(seconds))
		return gatewayError(c, 429, "rate_limited", "too many requests", nil)
	}
	c.Locals("routeName", route.Name)
	return s.proxy.Forward(c, route, actorID)
}
func (s *Server) ready(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), s.readinessTimeout)
	defer cancel()
	if err := s.proxy.Ready(ctx); err != nil {
		return c.Status(503).JSON(fiber.Map{"status": "unavailable"})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}
func (s *Server) requestContext(c *fiber.Ctx) error {
	id := validRequestID(c.Get("X-Request-ID"))
	if id == "" {
		raw := make([]byte, 16)
		_, _ = rand.Read(raw)
		id = hex.EncodeToString(raw)
	}
	c.Locals("requestId", id)
	c.Set("X-Request-ID", id)
	carrier := propagation.HeaderCarrier{}
	for key, values := range c.GetReqHeaders() {
		if len(values) > 0 {
			carrier.Set(key, values[0])
		}
	}
	ctx := otel.GetTextMapPropagator().Extract(c.UserContext(), carrier)
	ctx, span := otel.Tracer("api-gateway/http").Start(ctx, c.Method()+" "+c.Path())
	c.SetUserContext(ctx)
	started := time.Now()
	requests.Add(1)
	err := c.Next()
	status := c.Response().StatusCode()
	span.SetAttributes(attribute.Int("http.response.status_code", status))
	if err != nil || status >= 500 {
		span.SetStatus(codes.Error, stdhttp.StatusText(status))
	}
	span.End()
	s.logger.Info("gateway request", "request_id", id, "trace_id", span.SpanContext().TraceID().String(), "method", c.Method(), "path", c.Path(), "route", c.Locals("routeName"), "status", status, "duration_ms", time.Since(started).Milliseconds())
	return err
}
func validRequestID(value string) string {
	value = strings.TrimSpace(value)
	if len(value) == 0 || len(value) > 128 {
		return ""
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_' && r != '.' {
			return ""
		}
	}
	return value
}
func metrics(c *fiber.Ctx) error {
	c.Type("text/plain")
	return c.SendString(fmt.Sprintf("# TYPE gateway_http_requests_total counter\ngateway_http_requests_total %d\n", requests.Load()))
}
func (s *Server) errorHandler(c *fiber.Ctx, err error) error {
	status := 500
	if fiberErr := new(fiber.Error); errors.As(err, &fiberErr) {
		status = fiberErr.Code
	}
	message := "internal server error"
	if status < 500 {
		message = fiberErrMessage(status)
	}
	return gatewayError(c, status, httpCode(status), message, nil)
}
func fiberErrMessage(status int) string {
	if status == 413 {
		return "request body exceeded the configured limit"
	}
	return stdhttp.StatusText(status)
}
func httpCode(status int) string {
	switch status {
	case 400:
		return "invalid_request"
	case 401:
		return "unauthenticated"
	case 403:
		return "forbidden"
	case 404:
		return "not_found"
	case 405:
		return "method_not_allowed"
	case 413:
		return "request_too_large"
	case 415:
		return "unsupported_media_type"
	case 429:
		return "rate_limited"
	default:
		return "internal"
	}
}
func gatewayError(c *fiber.Ctx, status int, code, message string, fields map[string]string) error {
	id, _ := c.Locals("requestId").(string)
	return c.Status(status).JSON(gateway.ErrorEnvelope{Error: gateway.ErrorDetail{Code: code, Message: message, RequestID: id, Fields: fields}})
}
