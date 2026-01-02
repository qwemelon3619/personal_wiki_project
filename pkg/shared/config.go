// pkg/shared/config.go
package shared

import "time"

const (
	// DefaultCacheTTL provides a global default expiration for Redis keys (photo metadata).
	DefaultCacheTTL = 1 * time.Hour

	// ShortCacheTTL can be used for volatile data or frequently changing lists.
	ShortCacheTTL = 5 * time.Minute

	// PhotoCacheTTL is optimized for photo metadata caching.
	PhotoCacheTTL = 30 * time.Minute

	// MaxUploadFileSize is the maximum allowed file upload size (32MB).
	MaxUploadFileSize = 32 << 20

	// DefaultPhotoContainerName is the Azure Blob Storage container for photos.
	DefaultPhotoContainerName = "photos"

	// DefaultPhotoDatabase is the MongoDB database for photo gallery.
	DefaultPhotoDatabase = "PhotoGalleryDB"
)
