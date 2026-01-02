// services/gallery-service/internal/repository/gallery_repository.go
package repository

import (
	"context"

	"seungpyolee.com/pkg/model"
)

type CosmosDBRepository interface {
	// Photo queries
	GetPhotoByID(ctx context.Context, photoID string) (model.Photo, error)
	GetPhotosByUserID(ctx context.Context, userID string) ([]model.Photo, error)
	GetPhotosByDateRange(ctx context.Context, userID string, startDate, endDate string) ([]model.Photo, error)
}

type RedisRepository interface {
	// Photo caching
	GetPhotoMetadata(ctx context.Context, photoID string) (*model.Photo, error)
	SetPhotoMetadata(ctx context.Context, photoID string, photo *model.Photo) error
	InvalidatePhotoCache(ctx context.Context, photoID string) error
	// Gallery list caching
	GetGalleryCache(ctx context.Context, userID string) ([]model.Photo, error)
	SetGalleryCache(ctx context.Context, userID string, photos []model.Photo) error
	InvalidateGalleryCache(ctx context.Context, userID string) error
}
type AzureBlobRepository interface {
	GetBlob(ctx context.Context, blobName string) ([]byte, error)
}
