package repository

// services/editor-service/internal/repository/interfaces.go
import (
	"context"

	"seungpyolee.com/pkg/model"
)

type CosmosDBRepository interface {
	FindArticleByTitle(ctx context.Context, title string) (model.Article, error)
	UpsertArticle(ctx context.Context, article model.Article) error

	SaveRevision(ctx context.Context, revision model.Revision) error
}

type RedisRepository interface {
	SetArticle(ctx context.Context, article *model.Article) error
	DeleteCache(ctx context.Context, key string) error
}
