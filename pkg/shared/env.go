// pkg/shared/cache_utils.go
package shared

import (
	"math/rand"
	"os"
	"strconv"
	"time"
)

func GetRedisTTL() time.Duration {
	if val, ok := os.LookupEnv("REDIS_TTL_SECONDS"); ok {
		if seconds, err := strconv.Atoi(val); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return DefaultCacheTTL
}

// GetJitteredTTL adds random noise to the base TTL to prevent simultaneous expiration.
func GetJitteredTTL(baseTTL time.Duration) time.Duration {
	// Add random variation between 0% and 10% of base TTL
	f := rand.Float64() * 0.1
	jitter := time.Duration(float64(baseTTL) * f)
	return GetRedisTTL() + jitter
}
