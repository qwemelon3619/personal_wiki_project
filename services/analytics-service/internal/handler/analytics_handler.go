package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"seungpyolee.com/services/analytics-service/internal/service"
)

// AnalyticsHandler handles analytics-related HTTP requests
type AnalyticsHandler struct {
	service service.AnalyticsService
}

func NewAnalyticsHandler(svc service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: svc}
}

// HealthCheckHandler provides a simple health check endpoint
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// RecordAPICall handles recording API call statistics
func (h *AnalyticsHandler) RecordAPICall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint string `json:"endpoint"`
		UserID   string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Endpoint == "" {
		http.Error(w, "Endpoint is required", http.StatusBadRequest)
		return
	}

	if err := h.service.RecordAPICall(r.Context(), req.Endpoint, req.UserID); err != nil {
		log.Printf("[Handler] Failed to record API call: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

// GetSummary returns overall analytics summary
func (h *AnalyticsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetSummaryStats(r.Context())
	if err != nil {
		log.Printf("[Handler] Failed to get summary: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetEndpointStats returns statistics for a specific endpoint
func (h *AnalyticsHandler) GetEndpointStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpoint := vars["endpoint"]

	stats, err := h.service.GetEndpointStats(r.Context(), endpoint)
	if err != nil {
		log.Printf("[Handler] Failed to get endpoint stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetUserStats returns statistics for a specific user
func (h *AnalyticsHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	stats, err := h.service.GetUserStats(r.Context(), userID)
	if err != nil {
		log.Printf("[Handler] Failed to get user stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
