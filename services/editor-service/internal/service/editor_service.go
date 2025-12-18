package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"seungpyolee.com/pkg/model"
	"seungpyolee.com/services/editor-service/internal/repository"
)

type EditorService interface {
	UpdateArticle(ctx context.Context, title, content, comment string) error
}

type editorServiceImpl struct {
	cosmosRepo repository.CosmosDBRepository
	redisRepo  repository.RedisRepository
}

func NewEditorService(cRepo repository.CosmosDBRepository, rRepo repository.RedisRepository) EditorService {
	return &editorServiceImpl{
		cosmosRepo: cRepo,
		redisRepo:  rRepo,
	}
}

func (s *editorServiceImpl) UpdateArticle(ctx context.Context, title, content, comment string) error {
	existing, err := s.cosmosRepo.FindArticleByTitle(ctx, title)

	newVersion := 1
	if err == nil && existing.Title != "" {
		newVersion = existing.Version + 1
	}

	now := time.Now()
	article := model.Article{
		Title:     title,
		Content:   content,
		Version:   newVersion,
		UpdatedAt: now,
	}

	if err := s.cosmosRepo.UpsertArticle(ctx, article); err != nil {
		log.Printf("[Service] Failed to upsert article: %v", err)
		return err
	}

	revision := model.Revision{
		RevisionID: uuid.New().String(),
		ArticleID:  title,
		Version:    newVersion,
		Content:    content,
		Comment:    comment,
		CreatedAt:  now,
	}

	if err := s.cosmosRepo.SaveRevision(ctx, revision); err != nil {
		log.Printf("[Service] Failed to Update : %v", err)
	}

	if err := s.redisRepo.DeleteCache(ctx, title); err != nil {
		log.Printf("[Service] failed to delete cache: %v", err)
	}

	log.Printf("[Service] Complete Edit Article: %s (v%d)", title, newVersion)
	return nil
}
