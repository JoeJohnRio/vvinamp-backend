package service

import (
	"context"
	dbModel "vvinamp/model"
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
func (h *historyService) ListRecentHistory(ctx context.Context, limit int) ([]*dbModel.ListeningHistory, error) {
	// history, err := h.Repository.GetRecentHistoryByUser(ctx, limit)
	// if err != nil {
	// 	return nil, err
	// }

	// for _, item := range history {
	// 	// Fetch user (optional if you want it populated here)
	// 	// user, _ := h.userService.GetByID(ctx, item.UserID)
	// 	// item.User = user

	// 	// Note: You may or may not populate content here.
	// 	// If your GraphQL schema has field resolvers, you can defer this.
	// 	log.Println(item)
	// }
	return h.Repository.GetRecentHistoryByUser(ctx, limit)

	// panic("ListRecentHistory History service impl")
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
