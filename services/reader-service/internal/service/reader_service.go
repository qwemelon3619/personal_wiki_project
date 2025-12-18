// services/reader-service/internal/service/reader_service.go
package service

import (
	"context"
	"log"

	"golang.org/x/sync/singleflight"
	"seungpyolee.com/pkg/model"
	"seungpyolee.com/services/reader-service/internal/repository"
)

type ReaderService struct {
	dbRepo     repository.CosmosDBRepository // DB Implementation
	cacheRepo  repository.RedisRepository
	requestGrp singleflight.Group
}

func NewReaderService(dbRepo repository.CosmosDBRepository, cacheRepo repository.RedisRepository) *ReaderService {
	return &ReaderService{
		dbRepo:    dbRepo,
		cacheRepo: cacheRepo,
		//requestGrp: singleflight.Group{},
	}
}
func (s *ReaderService) GetArticle(ctx context.Context, title string) (*model.Article, error) {
	// 1. Primary Cache Lookup (
	article, err := s.cacheRepo.FindArticleByTitleInCache(ctx, title)
	if err == nil && article != nil {
		return article, nil
	}

	// 2. Singleflight to prevent cache stampede
	val, err, _ := s.requestGrp.Do(title, func() (interface{}, error) {
		//doucle check : if other goroutine may already populated the cache
		if cached, err := s.cacheRepo.FindArticleByTitleInCache(ctx, title); err == nil && cached != nil {
			return cached, nil
		}

		// DB Lookup
		dbArticle, err := s.dbRepo.FindArticleByTitleInDB(ctx, title)
		if err != nil {
			return nil, err
		}

		// Update Cache
		go func(a *model.Article) {
			if cacheErr := s.cacheRepo.SetArticleInCache(context.Background(), a); cacheErr != nil {
				log.Printf("Cache update failed for %s: %v", a.Title, cacheErr)
			}
		}(&dbArticle)

		return dbArticle, nil
	})

	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}

	return val.(*model.Article), nil
}
