package db

import (
	"context"
	"time"

	"github.com/nobletp001/sarvit/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect establishes a MongoDB client and verifies the connection.
// It returns (*mongo.Client, error) so callers can handle the error.
func Connect(cfg config.Config) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use Stable API v1 (recommended for MongoDB Atlas)
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	clientOpts := options.Client().
		ApplyURI(cfg.MongoURI).
		SetServerAPIOptions(serverAPI).
		SetConnectTimeout(30 * time.Second).
		SetSocketTimeout(30 * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	// Check the connection
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return client, nil
}

// GetCollection returns a specific MongoDB collection by name
func GetCollection(client *mongo.Client, dbName, collection string) *mongo.Collection {
	return client.Database(dbName).Collection(collection)
}
