package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"seungpyolee.com/pkg/auth"
	"seungpyolee.com/services/upload-service/internal/handler"
	"seungpyolee.com/services/upload-service/internal/repository"
	"seungpyolee.com/services/upload-service/internal/service"
)

func main() {
	// 1. Load environment variables
	cosmosURI := os.Getenv("COSMOS_URI")
	if cosmosURI == "" {
		cosmosURI = "mongodb://localhost:27017"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Azure Blob Storage connection string
	azureConnString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	if azureConnString == "" {
		log.Fatal("AZURE_STORAGE_CONNECTION_STRING environment variable is required")
	}

	azureContainerName := os.Getenv("AZURE_STORAGE_CONTAINER_NAME")
	if azureContainerName == "" {
		azureContainerName = "photos"
	}

	dbName := "PhotoGalleryDB"

	// 2. Initialize Repositories (Infrastructure Layer)
	log.Println("Initializing Repositories...")

	// MongoDB
	cosmosRepo := repository.NewCosmosDBRepository(cosmosURI, dbName)

	// Azure Blob Storage
	blobRepo, err := repository.NewAzureBlobRepository(azureConnString, azureContainerName)
	if err != nil {
		log.Fatalf("Failed to initialize Azure Blob Storage: %v", err)
	}

	// Redis
	redisRepo := repository.NewRedisRepository(redisAddr, "")

	// 3. Initialize Service Layer (Business Logic)
	log.Println("Initializing Service Layer...")
	uploaderSvc := service.NewUploaderService(cosmosRepo, blobRepo, redisRepo)

	// 4. Initialize Handler Layer (Transport Layer)
	log.Println("Initializing Handler Layer...")
	uploaderHandler := handler.NewUploaderHandler(uploaderSvc, service.NewAnalyticsClient("http://localhost:8082"))

	// 5. Configure Routes
	mux := http.NewServeMux()

	// Photo upload endpoint
	mux.HandleFunc("POST /api/upload", uploaderHandler.HandleUploadPhoto)

	// Get user's photos
	mux.HandleFunc("GET /api/photos", uploaderHandler.HandleGetPhotosByUser)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Uploader Service is Healthy"))
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

	// 6. Start Server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handlerToServe,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Uploader Service starting on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
