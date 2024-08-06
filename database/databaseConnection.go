package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client is the MongoDB client instance
var Client *mongo.Client

// DBinstance initializes and returns a MongoDB client instance
func DBinstance() *mongo.Client {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the MongoDB URL from the environment variable
	mongoDbURL := os.Getenv("MONGODB_URL")
	if mongoDbURL == "" {
		log.Fatal("MONGODB_URL not found in the environment variables")
	}

	// Create the client and connect to the server
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoDbURL))
	if err != nil {
		log.Fatal(err)
	}

	// Ping the primary to verify the connection is established
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB")

	return client
}

// Initialize the client variable
func init() {
	Client = DBinstance()
}

// OpenCollection returns a MongoDB collection from the database
func OpenCollection(collectionName string) *mongo.Collection {
	return Client.Database("cluster0").Collection(collectionName)
}
