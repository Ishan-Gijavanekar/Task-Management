package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/config"
	"github.com/Ishan-Gijavanekar/user-service/internal/database"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	"github.com/Ishan-Gijavanekar/user-service/pkg/logger"

	grpcTransport "github.com/Ishan-Gijavanekar/user-service/internal/transport/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config ", err)
	}

	appLogger := logger.New(cfg.App.Enviornment)
	appLogger.Info("Starting application", "name", cfg.App.Name, "enviournment", cfg.App.Enviornment)

	mongoDb, err := database.NewMongoDB(
		cfg.MongoDb.URI,
		cfg.MongoDb.Database,
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

	userService := service.NewUserService(&userRepository)

	grpcServer := grpcTransport.NewServer(cfg.GRPC.Port, userService, appLogger)

	serverErrors := make(chan error, 1)

	go func() {
		if err := grpcServer.Start(); err != nil {
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
			"grpc server failed",
			"error",
			err,
		)

		stop()
	}

	appLogger.Info("Shutdown signal received")

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.Stop()

	if err := mongoDb.Close(shutdownContext); err != nil {
		appLogger.Info("Error while closing mongoDB ", "error", err.Error())
	}

	appLogger.Info("Application stopped")
}
