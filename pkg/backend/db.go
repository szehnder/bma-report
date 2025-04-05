package backend

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbClient *mongo.Client
var BmaDB *mongo.Database
var addressesCollection *mongo.Collection
var cachedBMAReportsCollection *mongo.Collection
var llmInstructionsCollection *mongo.Collection

func ConnectDB() error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get MongoDB URI from environment variable, default to localhost
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Ping to ensure connection
	err = dbClient.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	fmt.Println("Connected to MongoDB!")
	BmaDB = dbClient.Database("bma_db")

	// Initialize collections with indexes
	if err := initCollections(); err != nil {
		return fmt.Errorf("failed to initialize collections: %v", err)
	}

	addressesCollection = BmaDB.Collection("addresses")
	cachedBMAReportsCollection = BmaDB.Collection("cached_bma_reports")
	llmInstructionsCollection = BmaDB.Collection("llm_instructions")

	return nil
}

func initCollections() error {
	ctx := context.Background()

	// Initialize raw_page_data collection
	rawCol := BmaDB.Collection("raw_page_data")
	_, err := rawCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "propertyDetails.address", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create index on raw_page_data: %v", err)
	}

	// Initialize addresses collection
	addrCol := BmaDB.Collection("addresses")
	_, err = addrCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "addressStr", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create index on addresses: %v", err)
	}

	// Initialize cached_bma_reports collection
	cachedCol := BmaDB.Collection("cached_bma_reports")
	_, err = cachedCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "primaryAddressId", Value: 1},
			{Key: "comparisonAddressIds", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create index on cached_bma_reports: %v", err)
	}

	return nil
}
