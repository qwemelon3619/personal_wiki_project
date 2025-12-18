// services/reader-service/internal/repository/redis_repo_impl.go
package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"seungpyolee.com/pkg/model"
	"seungpyolee.com/pkg/shared"
)

type RedisRepoImpl struct {
	client *redis.Client
}

func NewRedisRepository(addr, password string) RedisRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &RedisRepoImpl{client: rdb}
}

// FindArticleByTitleInCache retrieves an article using its title as a key.
func (r *RedisRepoImpl) FindArticleByTitleInCache(ctx context.Context, title string) (*model.Article, error) {
	key := fmt.Sprintf("article:title:%s", title)
	return r.getArticle(ctx, key)
}

// FindArticleByArticleIdInCache (Optional but recommended for consistency)
func (r *RedisRepoImpl) FindArticleByArticleIdInCache(ctx context.Context, articleId string) (*model.Article, error) {
	key := fmt.Sprintf("article:id:%s", articleId)
	return r.getArticle(ctx, key)
}

// SetArticleInCache stores an article with a jittered TTL.
func (r *RedisRepoImpl) SetArticleInCache(ctx context.Context, article *model.Article) error {
	key := fmt.Sprintf("article:title:%s", article.Title)
	data, err := json.Marshal(article)
	if err != nil {
		return err
	}

	ttl := shared.GetJitteredTTL(shared.DefaultCacheTTL)
	return r.client.Set(ctx, key, data, ttl).Err()
}

// FindRevisionByRevisionIdInCache retrieves a specific revision.
func (r *RedisRepoImpl) FindRevisionByRevisionIdInCache(ctx context.Context, revisionId string) (*model.Revision, error) {
	key := fmt.Sprintf("revision:%s", revisionId)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var revision model.Revision
	if err := json.Unmarshal([]byte(val), &revision); err != nil {
		return nil, err
	}
	return &revision, nil
}

// SetRevisionInCache stores a revision snapshot.
func (r *RedisRepoImpl) SetRevisionInCache(ctx context.Context, revision *model.Revision) error {
	key := fmt.Sprintf("revision:%s", revision.RevisionID)
	data, err := json.Marshal(revision)
	if err != nil {
		return err
	}

	ttl := shared.GetJitteredTTL(shared.DefaultCacheTTL)
	return r.client.Set(ctx, key, data, ttl).Err()
}

// FindRevisionsByArticleIdInCache retrieves a list of revisions.
// Uses a simple key for demonstration; for production, consider Redis Lists or Sets.
func (r *RedisRepoImpl) FindRevisionsByArticleIdInCache(ctx context.Context, articleId string) ([]model.Revision, error) {
	key := fmt.Sprintf("revisions:article:%s", articleId)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var revisions []model.Revision
	if err := json.Unmarshal([]byte(val), &revisions); err != nil {
		return nil, err
	}
	return revisions, nil
}

// Internal helper to avoid code duplication
func (r *RedisRepoImpl) getArticle(ctx context.Context, key string) (*model.Article, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var article model.Article
	if err := json.Unmarshal([]byte(val), &article); err != nil {
		return nil, err
	}
	return &article, nil
}
