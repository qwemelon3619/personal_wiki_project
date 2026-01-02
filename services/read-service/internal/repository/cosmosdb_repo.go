// services/gallery-service/internal/repository/cosmosdb_repo.go
package repository

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"seungpyolee.com/pkg/model"
)

type CosmosDBRepoImpl struct {
	client    *mongo.Client
	photoColl *mongo.Collection
}

func NewCosmosDBRepository(uri, dbName string) *CosmosDBRepoImpl {
	// 1. setup client options
	clientOptions := options.Client().ApplyURI(uri)
	// 2. create client
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect with CosmosDb: %v", err)
	}
	// 3. check actual connection (Ping)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to connect with CosmosDb(Ping): %v", err)
	}

	log.Println("Connected to CosmosDb successfully")

	db := client.Database(dbName)
	photoColl := db.Collection("photos")

	return &CosmosDBRepoImpl{
		client:    client,
		photoColl: photoColl,
	}
}

// GetPhotoByID retrieves a single photo by ID
func (r *CosmosDBRepoImpl) GetPhotoByID(ctx context.Context, photoID string) (model.Photo, error) {
	var photo model.Photo
	filter := bson.M{"_id": photoID}

	err := r.photoColl.FindOne(ctx, filter).Decode(&photo)
	if err == mongo.ErrNoDocuments {
		return model.Photo{}, nil
	}
	if err != nil {
		log.Printf("[Cosmos] Error fetching photo %s: %v", photoID, err)
		return model.Photo{}, err
	}
	return photo, nil
}

// GetPhotosByUserID retrieves all photos for a specific user (sorted by upload date descending)
func (r *CosmosDBRepoImpl) GetPhotosByUserID(ctx context.Context, userID string) ([]model.Photo, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.M{"uploaded_at": -1})

	cursor, err := r.photoColl.Find(ctx, filter, opts)
	if err != nil {
		log.Printf("[Cosmos] Error querying photos for user %s: %v", userID, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var photos []model.Photo
	if err := cursor.All(ctx, &photos); err != nil {
		log.Printf("[Cosmos] Error decoding photos: %v", err)
		return nil, err
	}

	return photos, nil
}

// GetPhotosByDateRange retrieves photos for a user within a date range
func (r *CosmosDBRepoImpl) GetPhotosByDateRange(ctx context.Context, userID string, startDate, endDate string) ([]model.Photo, error) {
	startTime, _ := time.Parse(time.RFC3339, startDate)
	endTime, _ := time.Parse(time.RFC3339, endDate)

	filter := bson.M{
		"user_id": userID,
		"uploaded_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}
	opts := options.Find().SetSort(bson.M{"uploaded_at": -1})

	cursor, err := r.photoColl.Find(ctx, filter, opts)
	if err != nil {
		log.Printf("[Cosmos] Error querying photos in date range: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var photos []model.Photo
	if err := cursor.All(ctx, &photos); err != nil {
		return nil, err
	}

	return photos, nil
}
