package repository

import (
	"context"
	"errors"
	"vvinamp/constant"
	"vvinamp/database/postgre"
	"vvinamp/model"
	"vvinamp/package/history"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type sqlRepository struct {
	logger *zap.Logger
}

// GetRecentHistoryByUser implements history.Repository.
func (s *sqlRepository) GetRecentHistoryByUser(ctx context.Context, limit int) ([]*model.ListeningHistory, error) {
	var history []*model.ListeningHistory
	userID, ok := ctx.Value(constant.Session).(string)
	if !ok || userID == "" {
		return nil, errors.New("unauthorized: no valid session 4")
	}
	err := s.DB(ctx).Where("user_id = ?", userID).Limit(limit).Order("played_at DESC").Find(&history).Error

	return history, err
}

type SQLRepositoryTarget struct {
	fx.In
	Connection *postgre.Connection
	Logger     *zap.Logger
}

func NewSQLRepository(target SQLRepositoryTarget) (history.Repository, error) {
	db, err := target.Connection.Client() // <-- capture both values

	if err != nil {
		return nil, err
	}

	if err := db.Debug().AutoMigrate(
		&model.ListeningHistory{},
		&model.PodcastSeries{},
		&model.PodcastEpisode{},
		&model.Playlist{},
	); err != nil {
		return nil, err
	}

	return &sqlRepository{
		logger: target.Logger,
	}, nil
}

func (m *sqlRepository) DB(ctx context.Context) *gorm.DB {
	return postgre.ForContext(ctx)
}
