package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	publicapi "github.com/Ishan-Gijavanekar/gateway-service/api"
	"github.com/Ishan-Gijavanekar/gateway-service/internal/auth"
	"github.com/Ishan-Gijavanekar/gateway-service/internal/config"
	"github.com/Ishan-Gijavanekar/gateway-service/internal/gateway"
	"github.com/Ishan-Gijavanekar/gateway-service/internal/observability"
	httptransport "github.com/Ishan-Gijavanekar/gateway-service/internal/transport/http"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "healthcheck" {
		healthcheck()
		return
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration invalid", "error", err)
		os.Exit(1)
	}

	root, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	shutdownTracing, err := observability.SetupTracing(root)
	if err != nil {
		logger.Error("tracing setup failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		ctx, done := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer done()
		_ = shutdownTracing(ctx)
	}()

	client := &http.Client{Transport: upstreamTransport(cfg.UpstreamTimeout)}
	proxy, err := gateway.NewProxy(client, cfg.UserServiceURL, cfg.TaskServiceURL, cfg.UpstreamTimeout, cfg.ResponseLimit)
	if err != nil {
		logger.Error("proxy setup failed", "error", err)
		os.Exit(1)
	}
	server, err := httptransport.New(cfg, proxy, auth.New(cfg.JWTKeys, cfg.JWTIssuer), logger, publicapi.OpenAPI)
	if err != nil {
		logger.Error("server setup failed", "error", err)
		os.Exit(1)
	}

	stopped := make(chan error, 1)
	go func() { stopped <- server.Start() }()
	logger.Info("api gateway started", "port", cfg.HTTPPort)
	select {
	case <-root.Done():
	case err = <-stopped:
		if err != nil && !errors.Is(err, fiber.ErrServiceUnavailable) {
			logger.Error("api gateway stopped unexpectedly", "error", err)
		}
		cancel()
	}

	ctx, done := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer done()
	if err = server.Stop(ctx); err != nil {
		logger.Error("api gateway shutdown failed", "error", err)
	}
	logger.Info("api gateway stopped")
}

func healthcheck() {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8088"
	}
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get("http://127.0.0.1:" + port + "/livez")
	if err != nil || response.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	_ = response.Body.Close()
}

func upstreamTransport(timeout time.Duration) *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		MaxConnsPerHost:       50,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: time.Second,
	}
}
