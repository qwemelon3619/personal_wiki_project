// services/editor-service/internal/repository/redis_repo.go
package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"seungpyolee.com/pkg/model"
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

func (r *RedisRepoImpl) SetArticle(ctx context.Context, article *model.Article) error {
	data, err := json.Marshal(article)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, "wiki:article:"+article.Title, data, 30*time.Minute).Err()
}

func (r *RedisRepoImpl) DeleteCache(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
