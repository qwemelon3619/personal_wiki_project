package service

import (
	"fmt"
	"log"
	"time"

	"seungpyolee.com/pkg/auth"
	"seungpyolee.com/pkg/model"
	"seungpyolee.com/services/auth-service/internal/repository"
)

// AuthService defines the interface for authentication business logic
type AuthService interface {
	Login(userID string) (string, error)
	Register(userID string) error
	Logout(tokenStr string) error
}

// AuthServiceImpl implements AuthService interface
type AuthServiceImpl struct {
	tokenManager     *auth.TokenManager
	userRepo         *repository.UserRepo
	fallbackUsersMap map[string]struct{} // For fallback when MongoDB is unavailable
}

// NewAuthService creates a new AuthService instance
func NewAuthService(tokenManager *auth.TokenManager, userRepo *repository.UserRepo) AuthService {
	return &AuthServiceImpl{
		tokenManager:     tokenManager,
		userRepo:         userRepo,
		fallbackUsersMap: make(map[string]struct{}),
	}
}

// Login handles user authentication and token generation
func (s *AuthServiceImpl) Login(userID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("invalid user ID")
	}

	// Legacy login - just issue token for any user ID
	accessToken, err := s.tokenManager.IssueAccessToken(userID, 15*time.Minute, nil)
	if err != nil {
		log.Printf("[AuthService] Failed to issue token for user %s: %v", userID, err)
		return "", err
	}

	log.Printf("[AuthService] Token issued for user: %s", userID)
	return accessToken, nil
}

// Register handles user registration
func (s *AuthServiceImpl) Register(userID string) error {
	if userID == "" {
		return fmt.Errorf("invalid user ID")
	}

	// Try MongoDB first if available
	if s.userRepo != nil {
		// Check if user already exists
		if exists, err := s.userRepo.GetUserByUserId(userID); err == nil && exists != nil {
			return fmt.Errorf("user already exists")
		}

		// Create new user in MongoDB
		newUser := model.User{
			UserID:    userID,
			CreatedAt: time.Now(),
		}
		if err := s.userRepo.CreateUser(newUser); err != nil {
			log.Printf("[AuthService] Failed to create user %s in MongoDB: %v", userID, err)
			return fmt.Errorf("failed to create user: %w", err)
		}
		log.Printf("[AuthService] User registered in MongoDB: %s", userID)
	} else {
		// Fallback to in-memory storage
		if _, exists := s.fallbackUsersMap[userID]; exists {
			return fmt.Errorf("user already exists")
		}
		s.fallbackUsersMap[userID] = struct{}{}
		log.Printf("[AuthService] User registered in memory: %s", userID)
	}

	return nil
}

// Logout handles token revocation
func (s *AuthServiceImpl) Logout(tokenStr string) error {
	if tokenStr == "" {
		return fmt.Errorf("invalid token")
	}

	// Verify and extract claims from token
	claims, err := s.tokenManager.VerifyToken(tokenStr)
	if err != nil {
		log.Printf("[AuthService] Token verification failed: %v", err)
		return fmt.Errorf("invalid token")
	}

	// Extract JTI and expiration
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return fmt.Errorf("invalid token claims: missing JTI")
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("invalid token claims: missing exp")
	}

	exp := time.Unix(int64(expFloat), 0)
	ttl := time.Until(exp)
	if ttl <= 0 {
		return fmt.Errorf("token already expired")
	}

	// Revoke token
	auth.RevokeJTI(jti, ttl)
	log.Printf("[AuthService] Token revoked: JTI=%s", jti)
	return nil
}
