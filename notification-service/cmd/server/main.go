package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ishan-Gijavanekar/notification-service/internal/config"

	"github.com/Ishan-Gijavanekar/notification-service/internal/database"

	"github.com/Ishan-Gijavanekar/notification-service/internal/messaging"

	httptransport "github.com/Ishan-Gijavanekar/notification-service/internal/transport/http"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			nil,
		),
	)

	if err := run(logger); err != nil {
		logger.Error(
			"notification service stopped with error",
			"error", err,
		)

		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {

	// Loading the configuration
	cfg, err := config.Load()

	if err != nil {
		return fmt.Errorf(
			"load configuration: %w",
			err,
		)
	}

	// Create the application lifecycle context.
	root, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// 3. Create a bounded startup context.
	startupCtx, startupCancel := context.WithTimeout(
		root,
		cfg.StartupTimeout,
	)
	defer startupCancel()

	// 4. Initialize MongoDB.
	mongoClient, err := database.Connect(
		cfg.MongoURI,
		startupCtx,
	)

	if err != nil {
		return err
	}

	defer func() {
		ctx, done := context.WithTimeout(
			context.Background(),
			cfg.ShutdownTimeout,
		)
		defer done()

		if err := mongoClient.Disconnect(ctx); err != nil {
			logger.Error(
				"mongodb disconnect failed",
				"error", err,
			)
		}
	}()

	logger.Info(
		"mongodb connected",
		"database", cfg.MongoDatabase,
	)

	//  Initialize NATS JetStream.
	natsClient, err := messaging.Connect(
		startupCtx,
		cfg.NATSURL,
		logger,
	)

	if err != nil {
		return err
	}

	defer func() {
		if err := natsClient.Close(); err != nil {
			logger.Error(
				"nats shutdown failed",
				"error", err,
			)
		}
	}()

	logger.Info("nats connected")

	// Initialize the domain event stream.
	err = messaging.EnsureDomainStream(
		startupCtx,
		natsClient.JetStream,
		cfg.NATSstream,
	)

	if err != nil {
		return err
	}

	logger.Info(
		"domain event stream initialized",
		"stream", cfg.NATSstream,
	)

	startupCancel()

	// 7. Initialize health checks.
	healthHandler := httptransport.NewHealthHandler(
		func(ctx context.Context) error {
			return database.Ping(ctx, mongoClient)
		},
		natsClient.Ping,
		cfg.DependencyTimeout,
	)

	// 8. Initialize the HTTP server.
	app := httptransport.NewServer(
		logger,
		healthHandler,
	)

	// 9. Start the HTTP server.
	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- app.Listen(
			":" + cfg.HttpPort,
		)
	}()

	logger.Info(
		"notification service starting",
		"http_port", cfg.HttpPort,
		"environment", cfg.Environment,
	)

	// 10. Wait for shutdown or server failure.
	var runErr error

	select {
	case <-root.Done():
		logger.Info("shutdown signal received")

	case err := <-serverErrors:
		if err != nil {
			runErr = fmt.Errorf(
				"http server failed: %w",
				err,
			)
		}
	}

	cancel()

	// 11. Gracefully stop the HTTP server.
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		runErr = errors.Join(
			runErr,
			fmt.Errorf(
				"http shutdown failed: %w",
				err,
			),
		)
	}

	logger.Info("notification service shutdown initiated")

	return runErr
}
