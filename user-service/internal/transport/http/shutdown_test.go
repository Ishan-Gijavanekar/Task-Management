package http

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestServerStopWithoutListener(t *testing.T) {
	server := &Server{app: fiber.New(), logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Stop(ctx); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}
}
