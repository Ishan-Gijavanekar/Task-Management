package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var ErrUpstreamUnavailable = errors.New("upstream unavailable")

type ErrorDetail struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"requestId"`
	Fields    map[string]string `json:"fields,omitempty"`
}
type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}

type Proxy struct {
	client           *http.Client
	userURL, taskURL *url.URL
	timeout          time.Duration
	responseLimit    int64
}

func NewProxy(client *http.Client, userURL, taskURL *url.URL, timeout time.Duration, responseLimit int64) (*Proxy, error) {
	if client == nil || userURL == nil || taskURL == nil || timeout <= 0 || responseLimit <= 0 {
		return nil, fmt.Errorf("client, upstream URLs, timeout, and response limit are required")
	}
	return &Proxy{client: client, userURL: userURL, taskURL: taskURL, timeout: timeout, responseLimit: responseLimit}, nil
}

func (p *Proxy) Forward(c *fiber.Ctx, route Route, actorID string) error {
	base := p.taskURL
	if route.Service == UserService {
		base = p.userURL
	}
	requestPath := c.Path()
	if route.ComposeUser {
		requestPath = "/api/v1/users/" + url.PathEscape(actorID)
	}
	target := *base
	target.Path = strings.TrimRight(base.Path, "/") + requestPath
	target.RawQuery = string(c.Request().URI().QueryString())
	ctx, cancel := context.WithTimeout(c.UserContext(), p.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, c.Method(), target.String(), bytes.NewReader(c.Body()))
	if err != nil {
		return writeGatewayError(c, 500, "internal", "internal server error", nil)
	}
	copyRequestHeaders(c, req)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	resp, err := p.client.Do(req)
	if err != nil {
		return writeGatewayError(c, 502, "upstream_unavailable", "service temporarily unavailable", nil)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, p.responseLimit+1))
	if err != nil {
		return writeGatewayError(c, 502, "invalid_upstream_response", "invalid service response", nil)
	}
	if int64(len(body)) > p.responseLimit {
		return writeGatewayError(c, 502, "upstream_response_too_large", "service response exceeded the configured limit", nil)
	}
	copyResponseHeaders(c, resp)
	if resp.StatusCode >= 400 {
		return normalizeUpstreamError(c, resp.StatusCode, body)
	}
	if route.ComposeUser {
		var user any
		if json.Unmarshal(body, &user) != nil {
			return writeGatewayError(c, 502, "invalid_upstream_response", "invalid service response", nil)
		}
		return c.Status(resp.StatusCode).JSON(fiber.Map{"user": user})
	}
	return c.Status(resp.StatusCode).Send(body)
}

func copyRequestHeaders(c *fiber.Ctx, req *http.Request) {
	for _, name := range []string{"Authorization", "Content-Type", "Accept", "Cookie", "Idempotency-Key", "User-Agent", "Traceparent", "Tracestate", "Baggage"} {
		if value := c.Get(name); value != "" {
			req.Header.Set(name, value)
		}
	}
	req.Header.Set("X-Request-ID", requestID(c))
	req.Header.Set("X-Forwarded-For", c.IP())
}
func copyResponseHeaders(c *fiber.Ctx, resp *http.Response) {
	for _, name := range []string{"Content-Type", "Cache-Control", "ETag", "Retry-After"} {
		if value := resp.Header.Get(name); value != "" {
			c.Set(name, value)
		}
	}
	for _, cookie := range resp.Header.Values("Set-Cookie") {
		c.Response().Header.Add("Set-Cookie", cookie)
	}
}
func normalizeUpstreamError(c *fiber.Ctx, status int, body []byte) error {
	var raw struct {
		Error struct {
			Code    string         `json:"code"`
			Message string         `json:"message"`
			Fields  map[string]any `json:"fields"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &raw)
	code := safeCode(raw.Error.Code)
	if code == "" {
		code = httpCode(status)
	}
	message := strings.TrimSpace(raw.Error.Message)
	if message == "" {
		message = http.StatusText(status)
	}
	fields := map[string]string{}
	for key, value := range raw.Error.Fields {
		if text, ok := value.(string); ok && key != "" && text != "" {
			fields[key] = text
		}
	}
	if len(fields) == 0 {
		fields = nil
	}
	return writeGatewayError(c, status, code, message, fields)
}
func safeCode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return ""
		}
	}
	return value
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
	case 409:
		return "conflict"
	case 410:
		return "gone"
	case 413:
		return "request_too_large"
	case 429:
		return "rate_limited"
	case 502, 503, 504:
		return "upstream_unavailable"
	default:
		return "internal"
	}
}
func writeGatewayError(c *fiber.Ctx, status int, code, message string, fields map[string]string) error {
	return c.Status(status).JSON(ErrorEnvelope{Error: ErrorDetail{Code: code, Message: message, RequestID: requestID(c), Fields: fields}})
}
func requestID(c *fiber.Ctx) string {
	if value, ok := c.Locals("requestId").(string); ok {
		return value
	}
	return c.GetRespHeader("X-Request-ID")
}
