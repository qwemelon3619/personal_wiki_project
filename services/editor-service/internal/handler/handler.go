// services/editor-service/internal/handler/admin_handler.go
package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"seungpyolee.com/services/editor-service/internal/service"
)

type EditorHandler struct {
	EditorService service.EditorService // 인터페이스에 의존
}

func NewEditorHandler(svc service.EditorService) *EditorHandler {
	return &EditorHandler{
		EditorService: svc,
	}
}

// services/editor-service/internal/handler/admin_handler.go 계속

func (h *EditorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 경로에서 제목 추출 (예: /api/edit/GoLanguage -> GoLanguage)
	title := strings.TrimPrefix(r.URL.Path, "/api/edit/")
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. 요청 바디 데이터 정의
	var req struct {
		Content string `json:"content"`
		Comment string `json:"comment"`
	}

	// 2. JSON 디코딩
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 3. 서비스 계층 호출
	ctx := r.Context()
	err := h.EditorService.UpdateArticle(ctx, title, req.Content, req.Comment)
	if err != nil {
		http.Error(w, "Failed to update article: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. 성공 응답
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Article updated successfully",
		"title":   title,
	})
}
