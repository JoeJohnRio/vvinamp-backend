package repository

import (
	"context"
	"vvinamp/database/postgre"
	dbModel "vvinamp/model"
	"vvinamp/package/song"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type sqlRepository struct {
	logger *zap.Logger
}

// GetTrackByID implements song.Repository.
func (m *sqlRepository) GetTrackByID(ctx context.Context, trackId string) (*dbModel.Track, error) {
	var track *dbModel.Track

	err := m.DB(ctx).Where("id = ?", trackId).First(&track).Error

	return track, err
}

type SQLRepositoryTarget struct {
	fx.In
	Connection *postgre.Connection
	Logger     *zap.Logger
}

func NewSQLRepository(target SQLRepositoryTarget) (song.Repository, error) {
	// db, err := target.Connection.Client() // <-- capture both values

	// if err != nil {
	// 	return nil, err
	// }

	// if err := db.Debug().AutoMigrate(&dbModel.Track{}); err != nil {
	// 	return nil, err
	// }

	return &sqlRepository{
		logger: target.Logger,
	}, nil
}

func (m *sqlRepository) DB(ctx context.Context) *gorm.DB {
	return postgre.ForContext(ctx)
}
