package main

import (
	"log"
	"net/http"
	"os"

	"seungpyolee.com/pkg/auth"
	"seungpyolee.com/services/auth-service/internal/handler"
	repo "seungpyolee.com/services/auth-service/internal/repository"
	"seungpyolee.com/services/auth-service/internal/service"
)

func main() {
	// Initialize token manager
	tm, err := auth.NewTokenManagerFromEnv()
	if err != nil {
		log.Fatalf("failed to init token manager: %v", err)
	}

	// Initialize MongoDB user repository
	dsn := os.Getenv("COSMOS_URI")
	if dsn == "" {
		dsn = "mongodb://mongodb:27017"
	}
	userRepo, err := repo.NewUserRepo(dsn)
	if err != nil {
		log.Printf("warning: failed to connect to MongoDB, falling back to in-memory users: %v", err)
	}

	// Initialize service and handler layers
	authService := service.NewAuthService(tm, userRepo)
	authHandler := handler.NewAuthHandler(authService)

	// Register HTTP endpoints
	http.HandleFunc("/auth/login", authHandler.HandleLogin)
	http.HandleFunc("/auth/register", authHandler.HandleRegister)
	http.HandleFunc("/auth/logout", authHandler.HandleLogout)
	http.HandleFunc("/auth/refresh", authHandler.HandleRefresh)

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	log.Printf("Auth service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
