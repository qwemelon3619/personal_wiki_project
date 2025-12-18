package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"seungpyolee.com/services/editor-service/internal/handler"
	"seungpyolee.com/services/editor-service/internal/repository"
	"seungpyolee.com/services/editor-service/internal/service"
)

func main() {
	// 1. 환경 변수 로드 (Azure 배포 환경 대비)
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

	// 2. Repository 초기화 (Infrastructure)
	log.Println("Initializing Repositories...")

	// Cosmos DB (MongoDB v2 Driver)
	cosmosRepo := repository.NewCosmosDBRepository(cosmosURI, dbName)

	// Redis
	redisRepo := repository.NewRedisRepository(redisAddr, "") // 비밀번호가 있다면 추가

	// 3. Service 초기화 및 의존성 주입 (Business Logic)
	// Service는 Repository 인터페이스들에 의존합니다.
	log.Println("Initializing Service Layer...")
	editorSvc := service.NewEditorService(cosmosRepo, redisRepo)

	// 4. Handler 초기화 및 의존성 주입 (Transport Layer)
	// Handler는 Service 인터페이스에 의존합니다.
	log.Println("Initializing Handler Layer...")
	editorHandler := handler.NewEditorHandler(editorSvc)

	// 5. 라우팅 설정
	mux := http.NewServeMux()

	// /api/edit/문서제목 경로로 들어오는 POST 요청 처리
	mux.Handle("/api/edit/", editorHandler)

	// 서버 헬스체크용
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Editor Service is Healthy"))
	})

	// 6. 서버 실행
	port := ":8080"
	server := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("🚀 Editor Service started on %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Critical Error: %v", err)
	}
}
