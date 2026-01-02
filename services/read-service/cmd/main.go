// services/gallery-service/main.go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"seungpyolee.com/pkg/auth"
	"seungpyolee.com/services/read-service/internal/handler"
	"seungpyolee.com/services/read-service/internal/repository"
	"seungpyolee.com/services/read-service/internal/service"
)

func main() {
	// 0. Initialize Repositories (DB and Cache)
	cosmosURI := os.Getenv("COSMOS_URI")
	if cosmosURI == "" {
		cosmosURI = "mongodb://localhost:27017"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	dbName := "PhotoGalleryDB"

	// 1. Initialize Repository (Infrastructure)
	log.Println("Initializing Repositories...")
	cosmosRepo := repository.NewCosmosDBRepository(cosmosURI, dbName)
	redisRepo := repository.NewRedisRepository(redisAddr, "")

	// 2. Initialize Service Layer
	galleryService := service.NewGalleryService(cosmosRepo, redisRepo)

	// 3. Initialize Handler Layer
	galleryHandler := handler.NewGalleryHandler(galleryService, service.NewAnalyticsClient("http://localhost:8082"))

	// 4. Configure Routes (Go 1.22+ syntax)
	mux := http.NewServeMux()

	// Photo retrieval endpoints
	mux.HandleFunc("GET /api/gallery/photo/{photoId}", galleryHandler.GetPhoto)
	mux.HandleFunc("GET /api/gallery", galleryHandler.GetGallery)
	mux.HandleFunc("GET /api/gallery/date", galleryHandler.GetGalleryByDateRange)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Gallery Service is Healthy"))
	})

	// Initialize JWT middleware (falls back to X-User-ID)
	tm, err := auth.NewTokenManagerFromEnv()
	if err != nil {
		log.Printf("warning: auth token manager not configured: %v (requests will require X-User-ID)", err)
	}

	var handlerToServe http.Handler = mux
	if tm != nil {
		handlerToServe = tm.Middleware(mux)
	}

	// 5. Configure Server
	server := &http.Server{
		Addr:         ":8081",
		Handler:      handlerToServe,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Println("Gallery Service starting on :8081")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
