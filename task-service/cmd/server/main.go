package main

import (
	"context"
	"errors"
	taskv1 "github.com/Ishan-Gijavanekar/task-service/api/proto/task/v1"
	"github.com/Ishan-Gijavanekar/task-service/internal/auth"
	"github.com/Ishan-Gijavanekar/task-service/internal/config"
	"github.com/Ishan-Gijavanekar/task-service/internal/observability"
	"github.com/Ishan-Gijavanekar/task-service/internal/repository"
	"github.com/Ishan-Gijavanekar/task-service/internal/service"
	grpctransport "github.com/Ishan-Gijavanekar/task-service/internal/transport/grpc"
	httptransport "github.com/Ishan-Gijavanekar/task-service/internal/transport/http"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, e := config.Load()
	if e != nil {
		logger.Error("configuration invalid", "error", e)
		os.Exit(1)
	}
	otel.SetTextMapPropagator(propagation.TraceContext{})
	root, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	shutdownTracing, e := observability.SetupTracing(root)
	if e != nil {
		logger.Error("tracing setup failed", "error", e)
		os.Exit(1)
	}
	defer shutdownTracing(context.Background())
	client, e := mongo.Connect(root, options.Client().ApplyURI(cfg.MongoURI))
	if e != nil {
		logger.Error("mongo connect failed", "error", e)
		os.Exit(1)
	}
	defer client.Disconnect(context.Background())
	repo, e := repository.NewMongoRepository(client.Database(cfg.MongoDatabase))
	if e != nil {
		logger.Error("repository failed", "error", e)
		os.Exit(1)
	}
	idx, stop := context.WithTimeout(root, 10*time.Second)
	e = repo.EnsureIndexes(idx)
	stop()
	if e != nil {
		logger.Error("indexes failed", "error", e)
		os.Exit(1)
	}
	svc, e := service.New(repo, service.StaticIdentityValidator{})
	if e != nil {
		logger.Error("service failed", "error", e)
		os.Exit(1)
	}
	jwt := auth.New(cfg.JWTKeys, cfg.JWTIssuer)
	app := fiber.New(fiber.Config{DisableStartupMessage: true, BodyLimit: 2 * 1024 * 1024})
	app.Use(observability.HTTP(logger))
	app.Get("/livez", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) })
	app.Get("/readyz", func(c *fiber.Ctx) error {
		ctx, done := context.WithTimeout(c.UserContext(), cfg.DependencyTimeout)
		defer done()
		if e := repo.Ping(ctx); e != nil {
			return c.Status(503).JSON(fiber.Map{"status": "unavailable"})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/metrics", observability.Metrics)
	api := app.Group("/api/v1", jwt.Fiber())
	httptransport.NewHandler(svc).Register(api)
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(observability.Unary(logger, cfg.ConcurrencyLimit), jwt.Unary()))
	taskv1.RegisterTaskServiceServer(grpcServer, grpctransport.NewHandler(svc))
	listener, e := net.Listen("tcp", ":"+cfg.GRPCPort)
	if e != nil {
		logger.Error("grpc listen failed", "error", e)
		os.Exit(1)
	}
	go func() {
		if e := grpcServer.Serve(listener); e != nil && !errors.Is(e, grpc.ErrServerStopped) {
			logger.Error("grpc server failed", "error", e)
			cancel()
		}
	}()
	go func() {
		if e := app.Listen(":" + cfg.HTTPPort); e != nil {
			logger.Error("http server failed", "error", e)
			cancel()
		}
	}()
	logger.Info("task service started", "http_port", cfg.HTTPPort, "grpc_port", cfg.GRPCPort)
	<-root.Done()
	ctx, done := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer done()
	grpcDone := make(chan struct{})
	go func() { grpcServer.GracefulStop(); close(grpcDone) }()
	select {
	case <-grpcDone:
	case <-ctx.Done():
		grpcServer.Stop()
	}
	_ = app.ShutdownWithContext(ctx)
	logger.Info("task service stopped")
}
