package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"seungpyolee.com/pkg/model"
)

// AnalyticsRepository defines data access methods for analytics
type AnalyticsRepository interface {
	SaveAPICall(ctx context.Context, endpoint string, userID string) error
	GetSummaryStats(ctx context.Context) (*model.AnalyticsSummary, error)
	GetEndpointStats(ctx context.Context, endpoint string) (*model.EndpointStats, error)
	GetUserStats(ctx context.Context, userID string) (*model.UserStats, error)
}

type AnalyticsRepositoryImpl struct {
	collection *mongo.Collection
}

func NewAnalyticsRepository(mongoURI, dbName, collectionName string) (*AnalyticsRepositoryImpl, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}
	coll := client.Database(dbName).Collection(collectionName)
	return &AnalyticsRepositoryImpl{collection: coll}, nil
}

// SaveAPICall saves an API call record to MongoDB
func (r *AnalyticsRepositoryImpl) SaveAPICall(ctx context.Context, endpoint string, userID string) error {
	log := model.APICallLog{
		Endpoint:  endpoint,
		UserID:    userID,
		Timestamp: time.Now(),
	}
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

// GetSummaryStats returns overall analytics summary
func (r *AnalyticsRepositoryImpl) GetSummaryStats(ctx context.Context) (*model.AnalyticsSummary, error) {
	// Total calls count
	totalCalls, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// Unique users count
	uniqueUsers, err := r.collection.Distinct(ctx, "userId", bson.M{})
	if err != nil {
		return nil, err
	}

	// Top endpoints (aggregation)
	pipeline := mongo.Pipeline{
		{{"$group", bson.D{{"_id", "$endpoint"}, {"count", bson.D{{"$sum", 1}}}}}},
		{{"$sort", bson.D{{"count", -1}}}},
		{{"$limit", 10}},
	}
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var topEndpoints []model.EndpointCount
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		topEndpoints = append(topEndpoints, model.EndpointCount{
			Endpoint: result.ID,
			Count:    result.Count,
		})
	}

	return &model.AnalyticsSummary{
		TotalCalls:   totalCalls,
		UniqueUsers:  int64(len(uniqueUsers)),
		TopEndpoints: topEndpoints,
	}, nil
}

// GetEndpointStats returns statistics for a specific endpoint
func (r *AnalyticsRepositoryImpl) GetEndpointStats(ctx context.Context, endpoint string) (*model.EndpointStats, error) {
	filter := bson.M{"endpoint": endpoint}

	// Total calls for this endpoint
	totalCalls, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Unique users for this endpoint
	uniqueUsers, err := r.collection.Distinct(ctx, "userId", filter)
	if err != nil {
		return nil, err
	}

	return &model.EndpointStats{
		Endpoint:     endpoint,
		TotalCalls:   totalCalls,
		UniqueUsers:  int64(len(uniqueUsers)),
		LastAccessed: time.Now(), // TODO: Get from actual last record
	}, nil
}

// GetUserStats returns statistics for a specific user
func (r *AnalyticsRepositoryImpl) GetUserStats(ctx context.Context, userID string) (*model.UserStats, error) {
	filter := bson.M{"userId": userID}

	// Total calls by this user
	totalCalls, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Top endpoints for this user
	pipeline := mongo.Pipeline{
		{{"$match", filter}},
		{{"$group", bson.D{{"_id", "$endpoint"}, {"count", bson.D{{"$sum", 1}}}}}},
		{{"$sort", bson.D{{"count", -1}}}},
		{{"$limit", 5}},
	}
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var topEndpoints []model.EndpointCount
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		topEndpoints = append(topEndpoints, model.EndpointCount{
			Endpoint: result.ID,
			Count:    result.Count,
		})
	}

	return &model.UserStats{
		UserID:       userID,
		TotalCalls:   totalCalls,
		TopEndpoints: topEndpoints,
	}, nil
}
