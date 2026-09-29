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
	"github.com/Ishan-Gijavanekar/user-service/pkg/logger"
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

	signalContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	appLogger.Info("Application started")
	<-signalContext.Done()

	appLogger.Info("Shutdown signal received")

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := mongoDb.Close(shutdownContext); err != nil {
		appLogger.Info("Error while closing mongoDB ", "error", err.Error())
	}

	appLogger.Info("Application stopped")
}
