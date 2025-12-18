// services/editor-service/internal/repository/cosmos_repo.go
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

// find
func (r *CosmosDBRepoImpl) FindArticleByTitle(ctx context.Context, title string) (model.Article, error) {
	var article model.Article
	err := r.articleColl.FindOne(ctx, bson.M{"_id": title}).Decode(&article)
	if err == mongo.ErrNoDocuments {
		return model.Article{}, nil
	}
	return article, err
}

// upsert article
func (r *CosmosDBRepoImpl) UpsertArticle(ctx context.Context, article model.Article) error {
	opts := options.UpdateOne().SetUpsert(true)
	filter := bson.M{"_id": article.Title}
	update := bson.M{
		"$set": bson.M{
			"content":    article.Content,
			"version":    article.Version,
			"updated_at": article.UpdatedAt,
		},
	}
	_, err := r.articleColl.UpdateOne(ctx, filter, update, opts)
	return err
}

// 수정 이력 저장
func (r *CosmosDBRepoImpl) SaveRevision(ctx context.Context, revision model.Revision) error {
	_, err := r.revisionColl.InsertOne(ctx, revision)
	return err
}
