package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	httpmiddleware "github.com/Ishan-Gijavanekar/user-service/internal/transport/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func newInfrastructureTestApp() *fiber.App {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := fiber.New(fiber.Config{
		ErrorHandler: NewErrorHandler(logger),
		BodyLimit:    1024,
	})
	app.Use(httpmiddleware.Security())
	app.Get("/health", health)
	app.Post("/body", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })
	return app
}

func TestHealthEndpoint(t *testing.T) {
	resp, err := newInfrastructureTestApp().Test(httptest.NewRequest("GET", "/health", nil))
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
}

func TestNotFoundResponse(t *testing.T) {
	resp, err := newInfrastructureTestApp().Test(httptest.NewRequest("GET", "/unknown", nil))
	if err != nil {
		t.Fatalf("app.Test() error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNotFound)
	}

	var body ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "NOT_FOUND" {
		t.Fatalf("error code = %q, want NOT_FOUND", body.Error.Code)
	}
}

func TestBodyLimit(t *testing.T) {
	req := httptest.NewRequest("POST", "/body", strings.NewReader(strings.Repeat("x", 1025)))
	if _, err := newInfrastructureTestApp().Test(req); err == nil {
		t.Fatal("app.Test() accepted a request larger than BodyLimit")
	}
}
