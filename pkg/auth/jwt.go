package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDContextKey = contextKey("userID")

// TokenManager holds keys and options for signing/verifying tokens
type TokenManager struct {
	signingMethod jwt.SigningMethod
	rsaPrivate    *rsa.PrivateKey
	rsaPublic     *rsa.PublicKey
	hmacSecret    []byte
	issuer        string
	audience      string
}

// NewTokenManagerFromEnv initializes TokenManager using env vars:
// JWT_ALG (RS256 or HS256), JWT_PRIVATE_KEY (PEM) or JWT_HMAC_SECRET, JWT_ISS, JWT_AUD
func NewTokenManagerFromEnv() (*TokenManager, error) {
	alg := os.Getenv("JWT_ALG")
	if alg == "" {
		alg = "HS256"
	}

	tm := &TokenManager{issuer: os.Getenv("JWT_ISS"), audience: os.Getenv("JWT_AUD")}

	switch alg {
	case "RS256":
		pemStr := os.Getenv("JWT_PRIVATE_KEY")
		if pemStr == "" {
			return nil, errors.New("JWT_PRIVATE_KEY required for RS256")
		}
		block, _ := pem.Decode([]byte(pemStr))
		if block == nil {
			return nil, errors.New("invalid PEM for JWT_PRIVATE_KEY")
		}
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			// try PKCS1
			keyRsa, err2 := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err2 != nil {
				return nil, fmt.Errorf("failed to parse RSA private key: %v/%v", err, err2)
			}
			tm.rsaPrivate = keyRsa
		} else {
			tm.rsaPrivate = key.(*rsa.PrivateKey)
		}
		tm.rsaPublic = &tm.rsaPrivate.PublicKey
		tm.signingMethod = jwt.GetSigningMethod("RS256")
	case "HS256":
		secret := os.Getenv("JWT_HMAC_SECRET")
		if secret == "" {
			return nil, errors.New("JWT_HMAC_SECRET required for HS256")
		}
		tm.hmacSecret = []byte(secret)
		tm.signingMethod = jwt.GetSigningMethod("HS256")
	default:
		return nil, fmt.Errorf("unsupported alg %s", alg)
	}

	return tm, nil
}

// IssueAccessToken issues an access token (short-lived)
func (tm *TokenManager) IssueAccessToken(userID string, ttl time.Duration, extra map[string]interface{}) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"iss": tm.issuer,
		"aud": tm.audience,
		"iat": now.Unix(),
		"exp": now.Add(ttl).Unix(),
		"jti": fmt.Sprintf("%d-%s", now.UnixNano(), userID),
	}
	for k, v := range extra {
		claims[k] = v
	}

	token := jwt.NewWithClaims(tm.signingMethod, claims)
	if tm.signingMethod.Alg() == "RS256" {
		return token.SignedString(tm.rsaPrivate)
	}
	return token.SignedString(tm.hmacSecret)
}

// VerifyToken verifies and returns claims
func (tm *TokenManager) VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if tm.signingMethod.Alg() == "RS256" {
			return tm.rsaPublic, nil
		}
		return tm.hmacSecret, nil
	}

	token, err := jwt.Parse(tokenStr, keyFunc, jwt.WithIssuer(tm.issuer), jwt.WithAudience(tm.audience))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	// Check if jti is revoked
	if jti, _ := claims["jti"].(string); jti != "" {
		if IsRevoked(jti) {
			return nil, errors.New("token revoked")
		}
	}

	return claims, nil
}

// Middleware validates Authorization: Bearer <token> and sets userID into context
func (tm *TokenManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr := strings.TrimPrefix(auth, "Bearer ")
			claims, err := tm.VerifyToken(tokenStr)
			if err == nil {
				if sub, _ := claims["sub"].(string); sub != "" {
					ctx := context.WithValue(r.Context(), UserIDContextKey, sub)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Fallback to X-User-ID for backward compatibility
		if uid := r.Header.Get("X-User-ID"); uid != "" {
			ctx := context.WithValue(r.Context(), UserIDContextKey, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		http.Error(w, "missing auth", http.StatusUnauthorized)
	})
}
