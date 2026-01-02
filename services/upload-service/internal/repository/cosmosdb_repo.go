// services/uploader-service/internal/repository/cosmosdb_repo.go
package repository

import (
	"context"
	"log"
	"time"

	"seungpyolee.com/pkg/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CosmosDBRepoImpl struct {
	client    *mongo.Client
	photoColl *mongo.Collection
	userColl  *mongo.Collection
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

	// Create indexes
	photoColl := db.Collection("photos")
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	}
	photoColl.Indexes().CreateOne(ctx, indexModel)

	userColl := db.Collection("users")

	return &CosmosDBRepoImpl{
		client:    client,
		photoColl: photoColl,
		userColl:  userColl,
	}
}

// SavePhoto inserts a new photo document
func (r *CosmosDBRepoImpl) SavePhoto(ctx context.Context, photo model.Photo) error {
	_, err := r.photoColl.InsertOne(ctx, photo)
	if err != nil {
		log.Printf("[Cosmos] Failed to save photo: %v", err)
		return err
	}
	log.Printf("[Cosmos] Photo saved: %s by user %s", photo.PhotoID, photo.UserID)
	return nil
}

// GetPhotosByUserID retrieves all photos for a specific user
func (r *CosmosDBRepoImpl) GetPhotosByUserID(ctx context.Context, userID string) ([]model.Photo, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.M{"uploaded_at": -1})

	cursor, err := r.photoColl.Find(ctx, filter, opts)
	if err != nil {
		log.Printf("[Cosmos] Failed to query photos for user %s: %v", userID, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var photos []model.Photo
	if err := cursor.All(ctx, &photos); err != nil {
		return nil, err
	}

	return photos, nil
}

// GetPhotoByID retrieves a single photo by ID
func (r *CosmosDBRepoImpl) GetPhotoByID(ctx context.Context, photoID string) (model.Photo, error) {
	var photo model.Photo
	err := r.photoColl.FindOne(ctx, bson.M{"_id": photoID}).Decode(&photo)
	if err == mongo.ErrNoDocuments {
		return model.Photo{}, nil
	}
	if err != nil {
		log.Printf("[Cosmos] Failed to get photo %s: %v", photoID, err)
		return model.Photo{}, err
	}
	return photo, nil
}

// UpdatePhotoMetadata updates EXIF and technical metadata for a photo
func (r *CosmosDBRepoImpl) UpdatePhotoMetadata(ctx context.Context, photoID string, metadata model.PhotoMetadata) error {
	filter := bson.M{"_id": photoID}
	update := bson.M{
		"$set": bson.M{
			"metadata": metadata,
		},
	}
	result, err := r.photoColl.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("[Cosmos] Failed to update metadata for photo %s: %v", photoID, err)
		return err
	}
	if result.MatchedCount == 0 {
		log.Printf("[Cosmos] Photo not found: %s", photoID)
	}
	return nil
}
