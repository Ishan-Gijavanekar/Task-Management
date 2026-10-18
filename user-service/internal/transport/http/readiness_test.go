package http

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type readinessStub struct{ err error }

func (s readinessStub) Ping(context.Context) error { return s.err }

func TestReadinessEndpoint(t *testing.T) {
	for _, test := range []struct {
		name   string
		check  ReadinessChecker
		status int
	}{
		{name: "ready", check: readinessStub{}, status: fiber.StatusOK},
		{name: "dependency unavailable", check: readinessStub{err: errors.New("unavailable")}, status: fiber.StatusServiceUnavailable},
		{name: "not configured", check: nil, status: fiber.StatusServiceUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/health/ready", readiness(test.check))
			response, err := app.Test(httptest.NewRequest("GET", "/health/ready", nil))
			if err != nil {
				t.Fatalf("app.Test() error: %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != test.status {
				t.Fatalf("status = %d, want %d", response.StatusCode, test.status)
			}
		})
	}
}
