package database

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoConfig struct {
	URI      string
	Database string
}

func ConnectMongo(ctx context.Context, config MongoConfig) (*mongo.Client, *mongo.Database, error) {
	if config.URI == "" {
		return nil, nil, errors.New("MongoDB URI is required")
	}
	if config.Database == "" {
		return nil, nil, errors.New("MongoDB database name is required")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.URI))
	if err != nil {
		return nil, nil, fmt.Errorf("connect to MongoDB: %w", err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, nil, fmt.Errorf("ping MongoDB: %w", err)
	}

	return client, client.Database(config.Database), nil
}
