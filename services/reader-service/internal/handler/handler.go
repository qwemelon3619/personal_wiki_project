package handler

// services/reader-service/internal/delivery/http/handler.go

import (
	"encoding/json"
	"net/http"

	"seungpyolee.com/pkg/model"
	"seungpyolee.com/services/reader-service/internal/service"
)

type ReaderHandler struct {
	readerService *service.ReaderService
}

func NewReaderHandler(s *service.ReaderService) *ReaderHandler {
	return &ReaderHandler{
		readerService: s,
	}
}

func (h *ReaderHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	title := r.PathValue("title")
	if title == "" {
		h.respondWithError(w, http.StatusBadRequest, "BAD_REQUEST", "Title is required", "Article")
		return
	}

	ctx := r.Context()
	article, err := h.readerService.GetArticle(ctx, title)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", "Article")
		return
	}

	if article == nil {
		h.respondWithError(w, http.StatusNotFound, "ARTICLE_NOT_FOUND", "Article not found", "Article")
		return
	}

	h.respondWithJSON(w, http.StatusOK, model.GenericResponse[model.Article]{
		Success: true,
		Data:    article,
		Error:   nil,
	})
}

// Helper: Standardized JSON Error Response
func (h *ReaderHandler) respondWithError(w http.ResponseWriter, httpStatus int, errorCode string, message string, target string) {
	h.respondWithJSON(w, httpStatus, model.ErrorDetail{
		Code:    errorCode,
		Message: message,
		Target:  target,
	})
}

// Helper: Standardized JSON Success Response
func (h *ReaderHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
