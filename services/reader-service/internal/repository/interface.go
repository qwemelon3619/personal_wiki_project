// services/reader-service/internal/repository/wiki_repository.go
package repository

import (
	"context"

	"seungpyolee.com/pkg/model"
)

type CosmosDBRepository interface {
	// Article related
	FindArticleByTitleInDB(ctx context.Context, title string) (model.Article, error)
	FindArticleByArticleIdInDB(ctx context.Context, articleId string) (model.Article, error)
	ListArticles(ctx context.Context, limit int, offset string) ([]model.Article, error) // For pagination

	// Revision related
	FindRevisionByArticleIdAndVersionInDB(ctx context.Context, articleId string, version int) (model.Revision, error) // Corrected signature
	FindRevisionsByArticleIdInDB(ctx context.Context, articleId string) ([]model.Revision, error)                     // Returns slice
}

type RedisRepository interface {
	// Article related
	FindArticleByTitleInCache(ctx context.Context, title string) (*model.Article, error)
	SetArticleInCache(ctx context.Context, article *model.Article) error // Required for Cache-Aside

	// Revision related
	FindRevisionByRevisionIdInCache(ctx context.Context, revisionId string) (*model.Revision, error)
	SetRevisionInCache(ctx context.Context, revision *model.Revision) error                          // Required for Cache-Aside
	FindRevisionsByArticleIdInCache(ctx context.Context, articleId string) ([]model.Revision, error) // Returns slice
}
