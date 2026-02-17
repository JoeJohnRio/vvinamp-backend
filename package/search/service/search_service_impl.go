package service

import (
	"context"
	"vvinamp/package/search"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Service struct {
	Repo search.Repository
}

type searchService struct {
	search.Repository
	logger *zap.Logger
}

// Search implements search.Service.
func (s *searchService) Search(ctx context.Context, query search.SearchQuery) (*search.SearchResult, error) {
	return s.Repository.Search(ctx, query)
}

type Target struct {
	fx.In
	Repository search.Repository
	Logger     *zap.Logger
}

func NewService(target Target) search.Service {
	return &searchService{
		Repository: target.Repository,
		logger:     target.Logger,
	}
}
