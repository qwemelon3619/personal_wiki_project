// services/reader-service/main.go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"seungpyolee.com/services/reader-service/internal/handler"
	"seungpyolee.com/services/reader-service/internal/repository"
	"seungpyolee.com/services/reader-service/internal/service"
)

func main() {
	// 0. Initialize Repositories (DB and Cache)
	cosmosURI := os.Getenv("COSMOS_URI")
	if cosmosURI == "" {
		// 로컬 테스트용 기본값 (환경에 맞게 수정하세요)
		cosmosURI = "mongodb://localhost:27017"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	dbName := "WikiDB"
	// cacheTTL := 1 * time.Hour

	// 2. initialize Repository (Infrastructure)
	log.Println("Initializing Repositories...")
	// Cosmos DB (MongoDB v2 Driver)
	cosmosRepo := repository.NewCosmosDBRepository(cosmosURI, dbName)
	// Redis
	redisRepo := repository.NewRedisRepository(redisAddr, "")
	readerService := service.NewReaderService(cosmosRepo, redisRepo)

	// 3. Handler 초기화 및 의존성 주입 (Transport Layer)
	handler := handler.NewReaderHandler(readerService)
	// 1. Create ServeMux
	mux := http.NewServeMux()

	// 2. Register Routes (Go 1.22 Syntax: {method} {path})
	mux.HandleFunc("GET /api/v1/wiki/{title}", handler.GetArticle)

	// 3. Configure Server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Println("Reader Service starting on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
