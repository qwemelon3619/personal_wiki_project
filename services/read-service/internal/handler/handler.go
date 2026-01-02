package handler

// services/gallery-service/internal/handler/gallery_handler.go

import (
	"encoding/json"
	"log"
	"net/http"

	"seungpyolee.com/services/read-service/internal/service"
)

type GalleryHandler struct {
	galleryService  *service.GalleryService
	analyticsClient *service.AnalyticsClient
}

func NewGalleryHandler(s *service.GalleryService, analyticsClient *service.AnalyticsClient) *GalleryHandler {
	return &GalleryHandler{
		galleryService:  s,
		analyticsClient: analyticsClient,
	}
}

// GetPhoto handles retrieval of a single photo by ID
// Expected header: "X-User-ID" for authorization
func (h *GalleryHandler) GetPhoto(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "X-User-ID header is required", http.StatusUnauthorized)
		return
	}

	photoID := r.PathValue("photoId")
	if photoID == "" {
		http.Error(w, "Photo ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	photo, err := h.galleryService.GetPhotoByID(ctx, photoID)
	if err != nil {
		log.Printf("[Handler] Error fetching photo: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if photo == nil {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	// Verify authorization: user can only see their own photos
	if photo.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(photo)

	// Record API call to analytics (async)
	if h.analyticsClient != nil {
		h.analyticsClient.RecordAPICall("/api/gallery/photo", userID)
	}
}

// GetGallery handles retrieval of all photos for the authenticated user
// Expected header: "X-User-ID"
func (h *GalleryHandler) GetGallery(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "X-User-ID header is required", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	photos, err := h.galleryService.GetPhotosByUser(ctx, userID)
	if err != nil {
		log.Printf("[Handler] Error fetching gallery: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"photos": photos,
		"count":  len(photos),
	})
}

// GetGalleryByDateRange handles filtered retrieval by date range
// Query params: startDate (RFC3339), endDate (RFC3339)
func (h *GalleryHandler) GetGalleryByDateRange(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "X-User-ID header is required", http.StatusUnauthorized)
		return
	}

	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")
	if startDate == "" || endDate == "" {
		http.Error(w, "startDate and endDate query params are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	photos, err := h.galleryService.GetPhotosByDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		log.Printf("[Handler] Error fetching photos by date: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"photos": photos,
		"count":  len(photos),
	})
}
