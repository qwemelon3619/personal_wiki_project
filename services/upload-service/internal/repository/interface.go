package repository

import (
	"context"
	"io"

	"seungpyolee.com/pkg/model"
)

// CosmosDBRepository handles photo metadata persistence in MongoDB
type CosmosDBRepository interface {
	SavePhoto(ctx context.Context, photo model.Photo) error
	GetPhotosByUserID(ctx context.Context, userID string) ([]model.Photo, error)
	GetPhotoByID(ctx context.Context, photoID string) (model.Photo, error)
	UpdatePhotoMetadata(ctx context.Context, photoID string, metadata model.PhotoMetadata) error
}

// AzureBlobRepository handles photo file storage in Azure Blob Storage
type AzureBlobRepository interface {
	UploadBlob(ctx context.Context, blobName string, fileData io.Reader, contentType string) (blobURL string, error error)
	DeleteBlob(ctx context.Context, blobName string) error
}

// RedisRepository handles caching of photo metadata
type RedisRepository interface {
	SetPhotoMetadata(ctx context.Context, photoID string, photo *model.Photo) error
	GetPhotoMetadata(ctx context.Context, photoID string) (*model.Photo, error)
	DeletePhotoCache(ctx context.Context, photoID string) error

	// Gallery list caching
	GetGalleryCache(ctx context.Context, userID string) ([]model.Photo, error)
	SetGalleryCache(ctx context.Context, userID string, photos []model.Photo) error
	InvalidateGalleryCache(ctx context.Context, userID string) error
}
