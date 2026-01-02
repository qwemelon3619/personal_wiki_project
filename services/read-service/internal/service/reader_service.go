// services/gallery-service/internal/service/gallery_service.go
package service

import (
	"context"
	"log"

	"golang.org/x/sync/singleflight"
	"seungpyolee.com/pkg/model"
	"seungpyolee.com/services/read-service/internal/repository"
)

type GalleryService struct {
	dbRepo     repository.CosmosDBRepository
	cacheRepo  repository.RedisRepository
	requestGrp singleflight.Group
}

func NewGalleryService(dbRepo repository.CosmosDBRepository, cacheRepo repository.RedisRepository) *GalleryService {
	return &GalleryService{
		dbRepo:    dbRepo,
		cacheRepo: cacheRepo,
	}
}

// GetPhotoByID retrieves a single photo with caching
func (s *GalleryService) GetPhotoByID(ctx context.Context, photoID string) (*model.Photo, error) {
	// 1. Try cache first
	photo, err := s.cacheRepo.GetPhotoMetadata(ctx, photoID)
	if err == nil && photo != nil {
		log.Printf("[Gallery] Cache hit for photo %s", photoID)
		return photo, nil
	}

	// 2. Singleflight to prevent cache stampede
	val, err, _ := s.requestGrp.Do(photoID, func() (interface{}, error) {
		// Double-check cache in case another goroutine populated it
		if cached, err := s.cacheRepo.GetPhotoMetadata(ctx, photoID); err == nil && cached != nil {
			return cached, nil
		}

		// DB lookup
		dbPhoto, err := s.dbRepo.GetPhotoByID(ctx, photoID)
		if err != nil {
			log.Printf("[Gallery] DB error for photo %s: %v", photoID, err)
			return nil, err
		}

		// Async cache update
		go func(p *model.Photo) {
			if cacheErr := s.cacheRepo.SetPhotoMetadata(context.Background(), p.PhotoID, p); cacheErr != nil {
				log.Printf("[Gallery] Cache update failed for photo %s: %v", p.PhotoID, cacheErr)
				// Non-fatal error
			}
		}(&dbPhoto)

		return &dbPhoto, nil
	})

	if err != nil {
		return nil, err
	}

	if val == nil {
		return nil, nil
	}

	return val.(*model.Photo), nil
}

// GetPhotosByUser retrieves all photos for a user with gallery list caching
func (s *GalleryService) GetPhotosByUser(ctx context.Context, userID string) ([]model.Photo, error) {
	// 1. Try gallery cache first
	photos, err := s.cacheRepo.GetGalleryCache(ctx, userID)
	if err == nil && photos != nil {
		log.Printf("[Gallery] Cache hit for gallery %s", userID)
		return photos, nil
	}

	// 2. Singleflight to prevent cache stampede
	val, err, _ := s.requestGrp.Do("gallery:"+userID, func() (interface{}, error) {
		// Double-check cache in case another goroutine populated it
		if cached, err := s.cacheRepo.GetGalleryCache(ctx, userID); err == nil && cached != nil {
			return cached, nil
		}

		// DB lookup
		dbPhotos, err := s.dbRepo.GetPhotosByUserID(ctx, userID)
		if err != nil {
			log.Printf("[Gallery] DB error for user %s: %v", userID, err)
			return nil, err
		}

		// Async cache update
		go func(photos []model.Photo) {
			if cacheErr := s.cacheRepo.SetGalleryCache(context.Background(), userID, photos); cacheErr != nil {
				log.Printf("[Gallery] Cache update failed for gallery %s: %v", userID, cacheErr)
				// Non-fatal error
			}
		}(dbPhotos)

		return dbPhotos, nil
	})

	if err != nil {
		return nil, err
	}

	if val == nil {
		return []model.Photo{}, nil
	}

	result := val.([]model.Photo)
	log.Printf("[Gallery] Retrieved %d photos for user %s", len(result), userID)
	return result, nil
}

// GetPhotosByDateRange retrieves photos within a date range
func (s *GalleryService) GetPhotosByDateRange(ctx context.Context, userID, startDate, endDate string) ([]model.Photo, error) {
	photos, err := s.dbRepo.GetPhotosByDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		log.Printf("[Gallery] Failed to fetch photos in date range: %v", err)
		return nil, err
	}

	return photos, nil
}
