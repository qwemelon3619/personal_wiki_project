// services/reader-service/internal/repository/mongodb_repo.go
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
	client       *mongo.Client
	articleColl  *mongo.Collection
	revisionColl *mongo.Collection
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

	return &CosmosDBRepoImpl{
		client:       client,
		articleColl:  db.Collection("articles"),
		revisionColl: db.Collection("revisions"),
	}
}

// FindArticleByTitleInDB retrieves an article by its title.
func (r *CosmosDBRepoImpl) FindArticleByTitleInDB(ctx context.Context, title string) (model.Article, error) {
	var article model.Article
	filter := bson.M{"title": title}

	err := r.articleColl.FindOne(ctx, filter).Decode(&article)
	if err != nil {
		return model.Article{}, err
	}
	return article, nil
}

// FindArticleByArticleIdInDB retrieves an article by its UUID string.
func (r *CosmosDBRepoImpl) FindArticleByArticleIdInDB(ctx context.Context, articleId string) (model.Article, error) {
	var article model.Article
	filter := bson.M{"_id": articleId}

	err := r.articleColl.FindOne(ctx, filter).Decode(&article)
	if err != nil {
		return model.Article{}, err
	}
	return article, nil
}

// ListArticles handles pagination using skip and limit.
func (r *CosmosDBRepoImpl) ListArticles(ctx context.Context, limit int, offset string) ([]model.Article, error) {
	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "updated_at", Value: -1}})

	// For simple implementation, offset is treated as skip count here.
	// In production, use cursor-based pagination with the 'offset' ID.
	cursor, err := r.articleColl.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []model.Article
	if err := cursor.All(ctx, &articles); err != nil {
		return nil, err
	}
	return articles, nil
}

// FindRevisionByArticleIdAndVersionInDB finds a specific version of an article.
func (r *CosmosDBRepoImpl) FindRevisionByArticleIdAndVersionInDB(ctx context.Context, articleId string, version int) (model.Revision, error) {
	var revision model.Revision
	filter := bson.M{
		"article_id": articleId,
		"version":    version,
	}

	err := r.revisionColl.FindOne(ctx, filter).Decode(&revision)
	if err != nil {
		return model.Revision{}, err
	}
	return revision, nil
}

// FindRevisionsByArticleIdInDB retrieves all history for a given article.
func (r *CosmosDBRepoImpl) FindRevisionsByArticleIdInDB(ctx context.Context, articleId string) ([]model.Revision, error) {
	opts := options.Find().SetSort(bson.D{{Key: "version", Value: -1}})
	filter := bson.M{"article_id": articleId}

	cursor, err := r.revisionColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var revisions []model.Revision
	if err := cursor.All(ctx, &revisions); err != nil {
		return nil, err
	}
	return revisions, nil
}
