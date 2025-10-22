package service

import (
	"context"
	"vvinamp/graphql/model"
	"vvinamp/package/history"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Service struct {
	Repo history.Repository
}

type historyService struct {
	history.Repository
	logger *zap.Logger
}

// ListRecentHistory implements history.Service.
func (h *historyService) ListRecentHistory(ctx context.Context, limit int) ([]*model.ListeningHistory, error) {
	return h.Repository.GetRecentHistoryByUser(ctx, limit)
}

type Target struct {
	fx.In
	Repository history.Repository
	Logger     *zap.Logger
}

func NewService(target Target) history.Service {
	return &historyService{
		Repository: target.Repository,
		logger:     target.Logger,
	}
}
