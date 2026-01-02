package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"seungpyolee.com/pkg/model"
)

// UserRepo kept name for compatibility; implementation uses MongoDB (CosmosDB)
type UserRepo struct {
	client     *mongo.Client
	collection *mongo.Collection
	timeout    time.Duration
}

// NewMySQLUserRepo accepts a MongoDB connection string (e.g. mongodb://mongodb:27017)
// and returns a repo that stores users in the "photoauth" DB, "users" collection.
func NewUserRepo(dsn string) (*UserRepo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(dsn)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("mongo connect error: %w", err)
	}

	// Ping to ensure connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping error: %w", err)
	}

	db := client.Database("photoauth")
	coll := db.Collection("users")

	// Ensure basic index on _id (default). Additional indexes (email unique) can be created here.
	// Example: create unique index on email if storing email field.

	return &UserRepo{
		client:     client,
		collection: coll,
		timeout:    10 * time.Second,
	}, nil
}

func (r *UserRepo) CreateUser(user model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetUserByUserId(userID string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}
func (r *UserRepo) GetUserByEmail(email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &user, nil
}

// Close disconnects the Mongo client
func (r *UserRepo) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}
