// pkg/shared/config.go
package shared

import "time"

const (
	// DefaultCacheTTL provides a global default expiration for Redis keys.
	DefaultCacheTTL = 1 * time.Hour

	// ShortCacheTTL can be used for volatile data.
	ShortCacheTTL = 5 * time.Minute
)
