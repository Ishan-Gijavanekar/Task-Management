package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Connect(uri string, ctx context.Context) (*mongo.Client, error) {
	clientOptions := options.Client().
		ApplyURI(uri).
		SetTimeout(5 * time.Second).
		SetServerSelectionTimeout(5 * time.Second).
		SetMaxPoolSize(50).
		SetMinPoolSize(0)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("error %v connecting to mongodb", err.Error())
	}

	if err := Ping(ctx, client); err != nil {
		cleanupContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = client.Disconnect(cleanupContext)

		return nil, fmt.Errorf("Error connectivity to mongodb")
	}

	return client, nil
}

func Ping(ctx context.Context, client *mongo.Client) error {
	return client.Ping(ctx, readpref.Primary())
}
