// services/uploader-service/internal/repository/redis_repo.go
package repository

import (
	"context"
	"encoding/json"
	"log"

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

// SetPhotoMetadata caches photo metadata with TTL
func (r *RedisRepoImpl) SetPhotoMetadata(ctx context.Context, photoID string, photo *model.Photo) error {
	data, err := json.Marshal(photo)
	if err != nil {
		log.Printf("[Redis] Failed to marshal photo: %v", err)
		return err
	}
	key := "photo:" + photoID
	err = r.client.Set(ctx, key, data, shared.PhotoCacheTTL).Err()
	if err != nil {
		log.Printf("[Redis] Failed to cache photo %s: %v", photoID, err)
	}
	return err
}

// GetPhotoMetadata retrieves cached photo metadata
func (r *RedisRepoImpl) GetPhotoMetadata(ctx context.Context, photoID string) (*model.Photo, error) {
	key := "photo:" + photoID
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		log.Printf("[Redis] Failed to get photo %s from cache: %v", photoID, err)
		return nil, err
	}

	var photo model.Photo
	if err := json.Unmarshal([]byte(val), &photo); err != nil {
		log.Printf("[Redis] Failed to unmarshal photo: %v", err)
		return nil, err
	}
	return &photo, nil
}

// DeletePhotoCache removes a photo from cache
func (r *RedisRepoImpl) DeletePhotoCache(ctx context.Context, photoID string) error {
	key := "photo:" + photoID
	err := r.client.Del(ctx, key).Err()
	if err != nil && err != redis.Nil {
		log.Printf("[Redis] Failed to delete cache for photo %s: %v", photoID, err)
	}
	return err
}

// GetGalleryCache retrieves cached gallery list for a user
func (r *RedisRepoImpl) GetGalleryCache(ctx context.Context, userID string) ([]model.Photo, error) {
	key := "gallery:" + userID
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		log.Printf("[Redis] Failed to get gallery cache for user %s: %v", userID, err)
		return nil, nil // Non-fatal
	}

	var photos []model.Photo
	if err := json.Unmarshal([]byte(val), &photos); err != nil {
		log.Printf("[Redis] Failed to unmarshal gallery: %v", err)
		return nil, nil
	}
	return photos, nil
}

// SetGalleryCache stores gallery list in cache with short TTL
func (r *RedisRepoImpl) SetGalleryCache(ctx context.Context, userID string, photos []model.Photo) error {
	data, err := json.Marshal(photos)
	if err != nil {
		log.Printf("[Redis] Failed to marshal gallery: %v", err)
		return err
	}
	key := "gallery:" + userID
	err = r.client.Set(ctx, key, data, shared.ShortCacheTTL).Err()
	if err != nil {
		log.Printf("[Redis] Failed to cache gallery for user %s: %v", userID, err)
	}
	return err
}

// InvalidateGalleryCache removes gallery cache for a user
func (r *RedisRepoImpl) InvalidateGalleryCache(ctx context.Context, userID string) error {
	key := "gallery:" + userID
	err := r.client.Del(ctx, key).Err()
	if err != nil && err != redis.Nil {
		log.Printf("[Redis] Failed to invalidate gallery cache for user %s: %v", userID, err)
	}
	return err
}
