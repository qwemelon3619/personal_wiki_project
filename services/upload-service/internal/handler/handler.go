// services/uploader-service/internal/handler/upload_handler.go
package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"seungpyolee.com/services/upload-service/internal/service"
)

type UploaderHandler struct {
	UploaderService service.UploaderService
	AnalyticsClient *service.AnalyticsClient
}

func NewUploaderHandler(svc service.UploaderService, analyticsClient *service.AnalyticsClient) *UploaderHandler {
	return &UploaderHandler{
		UploaderService: svc,
		AnalyticsClient: analyticsClient,
	}
}

// HandleUploadPhoto handles multipart form file uploads
// Expected form fields: "file", "title", "description"
// Expected header: "X-User-ID" for user identification
func (h *UploaderHandler) HandleUploadPhoto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from header (simplified auth - in production, use JWT)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "X-User-ID header is required", http.StatusUnauthorized)
		return
	}

	// Parse multipart form with 32MB max file size
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get file from form
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File not provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Call service to upload
	ctx := r.Context()
	photoID, err := h.UploaderService.UploadPhoto(ctx, userID, fileHeader.Filename, file)
	if err != nil {
		log.Printf("[Handler] Upload failed: %v", err)
		http.Error(w, "Failed to upload photo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"photoId": photoID,
		"message": "Photo uploaded successfully",
	})

	// Record API call to analytics (async)
	if h.AnalyticsClient != nil {
		h.AnalyticsClient.RecordAPICall("/api/upload", userID)
	}
}

// HandleGetPhotosByUser retrieves all photos for a user
func (h *UploaderHandler) HandleGetPhotosByUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "X-User-ID header is required", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	photos, err := h.UploaderService.GetPhotosByUser(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to fetch photos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"photos": photos,
		"count":  len(photos),
	})

	// Record API call to analytics (async)
	if h.AnalyticsClient != nil {
		h.AnalyticsClient.RecordAPICall("/api/photos", userID)
	}
}
