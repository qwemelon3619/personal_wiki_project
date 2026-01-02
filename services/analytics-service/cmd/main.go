package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"seungpyolee.com/services/analytics-service/internal/handler"
	"seungpyolee.com/services/analytics-service/internal/repository"
	"seungpyolee.com/services/analytics-service/internal/service"
)

func main() {
	r := mux.NewRouter()

	// MongoDB connection
	mongoURI := os.Getenv("COSMOS_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	dbName := "PhotoGalleryDB"
	collectionName := "api_call_logs"

	repo, err := repository.NewAnalyticsRepository(mongoURI, dbName, collectionName)
	if err != nil {
		log.Fatalf("[Analytics] MongoDB 연결 실패: %v", err)
	}

	svc := service.NewAnalyticsService(repo)
	analyticsHandler := handler.NewAnalyticsHandler(svc)

	r.HandleFunc("/health", handler.HealthCheckHandler).Methods("GET")

	r.HandleFunc("/api/analytics/record", analyticsHandler.RecordAPICall).Methods("POST")

	r.HandleFunc("/api/analytics/summary", analyticsHandler.GetSummary).Methods("GET")
	r.HandleFunc("/api/analytics/endpoint/{endpoint}", analyticsHandler.GetEndpointStats).Methods("GET")
	r.HandleFunc("/api/analytics/user/{userId}", analyticsHandler.GetUserStats).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	log.Printf("[Analytics] Service started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
