package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLiveness(t *testing.T) {
	handler := NewHealthHandler(
		healthyDependency,
		healthyDependency,
		time.Second,
	)

	app := NewServer(slog.Default(), handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/livez",
		nil,
	)

	response, err := app.Test(request)

	if err != nil {
		t.Fatal(err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected 200, received %d",
			response.StatusCode,
		)
	}
}

func TestReadinessHealthy(t *testing.T) {
	handler := NewHealthHandler(
		healthyDependency,
		healthyDependency,
		time.Second,
	)

	app := NewServer(slog.Default(), handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/readyz",
		nil,
	)

	response, err := app.Test(request)

	if err != nil {
		t.Fatal(err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected 200, received %d",
			response.StatusCode,
		)
	}
}

func TestReadinessUnavailable(t *testing.T) {
	handler := NewHealthHandler(
		unhealthyDependency,
		healthyDependency,
		time.Second,
	)

	app := NewServer(slog.Default(), handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/readyz",
		nil,
	)

	response, err := app.Test(request)

	if err != nil {
		t.Fatal(err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected 503, received %d",
			response.StatusCode,
		)
	}
}

func healthyDependency(context.Context) error {
	return nil
}

func unhealthyDependency(context.Context) error {
	return errors.New("dependency unavailable")
}
