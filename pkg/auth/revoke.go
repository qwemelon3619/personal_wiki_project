package auth

import (
	"sync"
	"time"
)

var (
	revokedMu sync.Mutex
	revoked   = map[string]time.Time{}
)

// RevokeJTI marks a jti as revoked until its expiry (or given ttl)
func RevokeJTI(jti string, ttl time.Duration) {
	revokedMu.Lock()
	revoked[jti] = time.Now().Add(ttl)
	revokedMu.Unlock()
}

// IsRevoked returns true if jti is revoked
func IsRevoked(jti string) bool {
	revokedMu.Lock()
	defer revokedMu.Unlock()
	exp, ok := revoked[jti]
	if !ok {
		return false
	}
	if time.Now().After(exp) {
		delete(revoked, jti)
		return false
	}
	return true
}
