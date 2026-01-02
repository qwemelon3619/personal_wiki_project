package service

import (
	"context"

	"seungpyolee.com/pkg/model"
	"seungpyolee.com/services/analytics-service/internal/repository"
)

// AnalyticsService defines business logic for analytics
type AnalyticsService interface {
	RecordAPICall(ctx context.Context, endpoint string, userID string) error
	GetSummaryStats(ctx context.Context) (*model.AnalyticsSummary, error)
	GetEndpointStats(ctx context.Context, endpoint string) (*model.EndpointStats, error)
	GetUserStats(ctx context.Context, userID string) (*model.UserStats, error)
}

type analyticsServiceImpl struct {
	repo repository.AnalyticsRepository
}

func NewAnalyticsService(repo repository.AnalyticsRepository) AnalyticsService {
	return &analyticsServiceImpl{repo: repo}
}

// RecordAPICall records an API call event
func (s *analyticsServiceImpl) RecordAPICall(ctx context.Context, endpoint string, userID string) error {
	return s.repo.SaveAPICall(ctx, endpoint, userID)
}

// GetSummaryStats returns overall statistics
func (s *analyticsServiceImpl) GetSummaryStats(ctx context.Context) (*model.AnalyticsSummary, error) {
	return s.repo.GetSummaryStats(ctx)
}

// GetEndpointStats returns statistics for a specific endpoint
func (s *analyticsServiceImpl) GetEndpointStats(ctx context.Context, endpoint string) (*model.EndpointStats, error) {
	return s.repo.GetEndpointStats(ctx, endpoint)
}

// GetUserStats returns statistics for a specific user
func (s *analyticsServiceImpl) GetUserStats(ctx context.Context, userID string) (*model.UserStats, error) {
	return s.repo.GetUserStats(ctx, userID)
}
