package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB(uri, databaseName string, logger slog.Logger) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(uri),
	)
	if err != nil {
		return nil, fmt.Errorf("Error connecting MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("Error in Ping %v", err)
	}

	logger.Info("mongodb connected", "database", databaseName)

	return &MongoDB{
		Client:   client,
		Database: client.Database(databaseName),
	}, nil
}

func (m *MongoDB) Close(ctx context.Context) error {
	if m.Client == nil {
		return nil
	}

	return m.Client.Disconnect(ctx)
}
