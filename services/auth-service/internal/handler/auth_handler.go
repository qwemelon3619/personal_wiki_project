package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"seungpyolee.com/services/auth-service/internal/service"
)

// AuthHandler handles HTTP requests for authentication endpoints
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// HandleLogin handles POST /auth/login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	accessToken, err := h.authService.Login(req.UserID)
	if err != nil {
		log.Printf("[Handler] Login failed for user %s: %v", req.UserID, err)
		http.Error(w, "failed to issue token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"accessToken": accessToken})
}

// HandleRegister handles POST /auth/register
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := h.authService.Register(req.UserID); err != nil {
		if strings.Contains(err.Error(), "exists") {
			http.Error(w, "user exists", http.StatusConflict)
			return
		}
		log.Printf("[Handler] Register failed for user %s: %v", req.UserID, err)
		http.Error(w, "failed to register", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

// HandleLogout handles POST /auth/logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing authorization", http.StatusBadRequest)
		return
	}

	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		http.Error(w, "bad authorization header", http.StatusBadRequest)
		return
	}

	tokenStr := authHeader[7:]
	if err := h.authService.Logout(tokenStr); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		log.Printf("[Handler] Logout failed: %v", err)
		http.Error(w, "logout failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"})
}

// HandleRefresh handles POST /auth/refresh (placeholder)
func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
