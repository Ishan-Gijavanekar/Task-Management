package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/Ishan-Gijavanekar/user-service/internal/config"
	"github.com/Ishan-Gijavanekar/user-service/internal/database"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	"github.com/Ishan-Gijavanekar/user-service/pkg/logger"

	grpcTransport "github.com/Ishan-Gijavanekar/user-service/internal/transport/grpc"
	httpTransport "github.com/Ishan-Gijavanekar/user-service/internal/transport/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config ", err)
	}

	appLogger := logger.New(cfg.App.Environment)
	appLogger.Info("Starting application", "name", cfg.App.Name, "environment", cfg.App.Environment)

	mongoDb, err := database.NewMongoDB(
		cfg.MongoDB.URI,
		cfg.MongoDB.Database,
		*appLogger,
	)
	if err != nil {
		appLogger.Error("Failed to connect to mongo", "error", err.Error())
		os.Exit(1)
	}

	mongoUserRepository := repository.NewMongoUserRepository(
		mongoDb.Database,
	)
	var userRepository repository.UserRespository = mongoUserRepository
	mongoTokenRepository := repository.NewMongoAuthTokenRepository(mongoDb.Database)

	indexContext, cancelIndexes := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancelIndexes()

	if err := mongoUserRepository.EnsureIndexes(indexContext); err != nil {
		appLogger.Error(
			"Failed to initialize user indexes",
			"error",
			err,
		)

		os.Exit(1)
	}
	if err := mongoTokenRepository.EnsureIndexes(indexContext); err != nil {
		appLogger.Error("Failed to initialize auth token indexes", "error", err)
		os.Exit(1)
	}

	userService := service.NewUserService(&userRepository)
	if err := userService.Validate(); err != nil {
		appLogger.Error("Invalid user service configuration", "error", err)
		os.Exit(1)
	}
	signingKeys := cfg.JWT.PreviousKeys
	signingKeys[cfg.JWT.ActiveKeyID] = cfg.JWT.Secret
	jwtManager := auth.NewRotatingJWTManager(
		cfg.JWT.ActiveKeyID,
		signingKeys,
		cfg.JWT.Issuer,
		cfg.JWT.AccessTokenDuration,
	)
	if err := jwtManager.Validate(); err != nil {
		appLogger.Error("Invalid JWT configuration", "error", err)
		os.Exit(1)
	}
	authOptions := []service.AuthServiceOption{
		service.WithAuthTokenRepository(mongoTokenRepository),
		service.WithRefreshTokenDuration(cfg.Auth.RefreshTokenDuration),
		service.WithEmailVerificationRequired(cfg.Auth.RequireEmailVerification),
	}
	if cfg.App.Environment == "development" {
		authOptions = append(authOptions, service.WithAuthNotificationSender(
			service.NewDevelopmentAuthNotifier(appLogger, cfg.Auth.FrontendURL),
		))
	} else if cfg.Auth.SMTPHost != "" {
		authOptions = append(authOptions, service.WithAuthNotificationSender(
			service.NewSMTPAuthNotifier(
				cfg.Auth.SMTPHost, cfg.Auth.SMTPPort, cfg.Auth.SMTPUsername,
				cfg.Auth.SMTPPassword, cfg.Auth.SMTPFrom, cfg.Auth.FrontendURL,
			),
		))
	}
	authService := service.NewAuthService(userRepository, jwtManager, authOptions...)
	if err := authService.Validate(); err != nil {
		appLogger.Error("Invalid authentication configuration", "error", err)
		os.Exit(1)
	}

	grpcServer := grpcTransport.NewServer(
		cfg.GRPC.Port,
		cfg.GRPC.EnableReflection,
		userService,
		authService,
		jwtManager,
		appLogger,
	)

	httpServer := httpTransport.NewServer(
		cfg.HTTP,
		cfg.Auth,
		userService,
		authService,
		jwtManager,
		appLogger,
		mongoDb,
	)
	if err := grpcServer.Validate(); err != nil {
		appLogger.Error("Invalid gRPC server configuration", "error", err)
		os.Exit(1)
	}
	if err := httpServer.Validate(); err != nil {
		appLogger.Error("Invalid HTTP server configuration", "error", err)
		os.Exit(1)
	}

	serverErrors := make(chan error, 2)

	go func() {
		if err := grpcServer.Start(); err != nil {
			serverErrors <- err
		}
	}()

	go func() {
		if err := httpServer.Start(); err != nil {
			serverErrors <- err
		}
	}()

	signalContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	appLogger.Info("Application started")
	select {

	case <-signalContext.Done():

		appLogger.Info(
			"shutdown signal received",
		)

	case err := <-serverErrors:

		appLogger.Error(
			"server failed",
			"error",
			err,
		)

		stop()
	}

	appLogger.Info("Shutdown signal received")

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.Stop()

	if err := httpServer.Stop(shutdownContext); err != nil {
		appLogger.Info("Error while stopping http server", "error", err.Error())
	}

	if err := mongoDb.Close(shutdownContext); err != nil {
		appLogger.Info("Error while closing mongoDB ", "error", err.Error())
	}

	appLogger.Info("Application stopped")
}
