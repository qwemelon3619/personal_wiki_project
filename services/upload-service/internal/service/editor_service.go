package service

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"seungpyolee.com/pkg/model"
	"seungpyolee.com/services/upload-service/internal/repository"
)

// UploaderService handles photo upload, EXIF extraction, and metadata storage
type UploaderService interface {
	UploadPhoto(ctx context.Context, userID string, fileName string, fileData io.Reader) (photoID string, err error)
	GetPhotosByUser(ctx context.Context, userID string) ([]model.Photo, error)
}

type uploaderServiceImpl struct {
	cosmosRepo    repository.CosmosDBRepository
	blobRepo      repository.AzureBlobRepository
	redisRepo     repository.RedisRepository
	exifExtractor *ExifExtractor
}

func NewUploaderService(
	cosmosRepo repository.CosmosDBRepository,
	blobRepo repository.AzureBlobRepository,
	redisRepo repository.RedisRepository,
) UploaderService {
	return &uploaderServiceImpl{
		cosmosRepo:    cosmosRepo,
		blobRepo:      blobRepo,
		redisRepo:     redisRepo,
		exifExtractor: NewExifExtractor(),
	}
}

// UploadPhoto orchestrates file upload, EXIF extraction, and metadata storage
func (s *uploaderServiceImpl) UploadPhoto(ctx context.Context, userID string, fileName string, fileData io.Reader) (string, error) {
	photoID := uuid.New().String()
	now := time.Now()

	// Read entire file into memory so we can upload multiple sizes and extract EXIF
	fileBytes, err := io.ReadAll(fileData)
	if err != nil {
		log.Printf("[Service] Failed to read uploaded file: %v", err)
		return "", err
	}

	// Determine content type and extension from filename
	ext := strings.ToLower(filepath.Ext(fileName))
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload original
	originalBlobName := fmt.Sprintf("%s/%s%s", userID, photoID, ext)
	if _, err := s.blobRepo.UploadBlob(ctx, originalBlobName, bytes.NewReader(fileBytes), contentType); err != nil {
		log.Printf("[Service] Failed to upload original blob: %v", err)
		return "", err
	}

	// 2. Extract EXIF metadata from buffer
	metadata := s.exifExtractor.ExtractMetadata(bytes.NewReader(fileBytes))

	// If content is an image, generate resized versions
	if strings.HasPrefix(contentType, "image/") {
		img, err := imaging.Decode(bytes.NewReader(fileBytes))
		if err != nil {
			log.Printf("[Service] Failed to decode image for resizing: %v", err)
		} else {
			sizes := []int{1080, 720, 480}
			for _, w := range sizes {
				// Only resize when image is wider than target width
				if img.Bounds().Dx() <= w {
					// Skip resizing if original is smaller or equal
					continue
				}
				resized := imaging.Resize(img, w, 0, imaging.Lanczos)

				// Encode resized image to JPEG
				var buf bytes.Buffer
				opt := &jpeg.Options{Quality: 85}
				if err := jpeg.Encode(&buf, resized, opt); err != nil {
					log.Printf("[Service] Failed to encode resized image %d: %v", w, err)
					continue
				}

				resizeBlobName := fmt.Sprintf("%s/%s_%d.jpg", userID, photoID, w)
				if _, err := s.blobRepo.UploadBlob(ctx, resizeBlobName, bytes.NewReader(buf.Bytes()), "image/jpeg"); err != nil {
					log.Printf("[Service] Failed to upload resized blob %d: %v", w, err)
					continue
				}
			}
		}
	}

	// 3. Create Photo document
	photo := model.Photo{
		PhotoID:    photoID,
		UserID:     userID,
		FileName:   fileName,
		MimeType:   contentType,
		UploadedAt: now,
		Metadata:   metadata,
	}

	// 4. Save to MongoDB
	if err := s.cosmosRepo.SavePhoto(ctx, photo); err != nil {
		log.Printf("[Service] Failed to save photo metadata: %v", err)
		// Clean up blobs (original + resized) if DB fails
		go func() {
			ctx := context.Background()
			// original
			if delErr := s.blobRepo.DeleteBlob(ctx, originalBlobName); delErr != nil {
				log.Printf("[Service] Failed to delete original blob during cleanup: %v", delErr)
			}
			// resized variants
			sizes := []int{1080, 720, 480}
			for _, w := range sizes {
				resizeBlobName := fmt.Sprintf("%s/%s_%d.jpg", userID, photoID, w)
				if delErr := s.blobRepo.DeleteBlob(ctx, resizeBlobName); delErr != nil {
					// non-fatal
					log.Printf("[Service] Failed to delete resized blob during cleanup: %v", delErr)
				}
			}
		}()
		return "", err
	}

	// 5. Cache the photo metadata
	if err := s.redisRepo.SetPhotoMetadata(ctx, photoID, &photo); err != nil {
		log.Printf("[Service] Failed to cache photo metadata: %v (non-fatal)", err)
		// Cache failure is non-fatal
	}

	// 6. Invalidate gallery cache for this user (since we added a new photo)
	if err := s.redisRepo.InvalidateGalleryCache(ctx, userID); err != nil {
		log.Printf("[Service] Failed to invalidate gallery cache: %v (non-fatal)", err)
		// Cache failure is non-fatal
	}

	log.Printf("[Service] Photo uploaded successfully: %s by user %s", photoID, userID)
	return photoID, nil
}

// GetPhotosByUser retrieves all photos for a given user
func (s *uploaderServiceImpl) GetPhotosByUser(ctx context.Context, userID string) ([]model.Photo, error) {
	photos, err := s.cosmosRepo.GetPhotosByUserID(ctx, userID)
	if err != nil {
		log.Printf("[Service] Failed to fetch photos for user %s: %v", userID, err)
		return nil, err
	}
	return photos, nil
}
