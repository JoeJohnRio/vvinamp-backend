package repository

import (
	"context"
	"vvinamp/database/postgre"
	dbModel "vvinamp/model"
	"vvinamp/model/dto"
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

// GetAlbumByID implements song.Repository.
func (m *sqlRepository) GetAlbumByID(ctx context.Context, albumId string) (*dbModel.Album, error) {
	var album *dbModel.Album

	err := m.DB(ctx).Where("id = ?", albumId).First(&album).Error

	return album, err
}

// GetTracksByAlbumID implements song.Repository.
func (m *sqlRepository) GetTracksByAlbumID(ctx context.Context, albumId string) ([]*dbModel.Track, error) {
	var tracks []*dbModel.Track

	err := m.DB(ctx).Where("album_id = ?", albumId).Order("title ASC").Find(&tracks).Error

	return tracks, err
}

// GetTracksByAlbumIDWithDetails implements song.Repository.
func (m *sqlRepository) GetTracksByAlbumIDWithDetails(ctx context.Context, albumId string) ([]*dto.TrackWithDetails, error) {
	var tracks []*dto.TrackWithDetails

	err := m.DB(ctx).
		Table("tracks").
		Select("tracks.*, albums.cover_art as album_cover_url, artists.name as artist_name").
		Joins("JOIN albums ON tracks.album_id = albums.id").
		Joins("JOIN artists ON albums.artist_id = artists.id").
		Where("tracks.album_id = ?", albumId).
		Order("tracks.title ASC").
		Scan(&tracks).Error

	return tracks, err
}

// GetTracksByGenres implements song.Repository.
func (m *sqlRepository) GetTracksByGenres(ctx context.Context, genreIds []string, limit int, excludeIds []string) ([]*dto.TrackWithDetails, error) {
	var tracks []*dto.TrackWithDetails

	query := m.DB(ctx).
		Table("tracks").
		Select("tracks.*, albums.cover_art as album_cover_url, artists.name as artist_name").
		Joins("JOIN albums ON tracks.album_id = albums.id").
		Joins("JOIN artists ON albums.artist_id = artists.id")

	if len(genreIds) > 0 {
		query = query.Where("albums.genre_id IN ?", genreIds)
	}

	if len(excludeIds) > 0 {
		query = query.Where("tracks.id NOT IN ?", excludeIds)
	}

	err := query.Limit(limit).Scan(&tracks).Error
	return tracks, err
}

// GetRandomTracks implements song.Repository.
func (m *sqlRepository) GetRandomTracks(ctx context.Context, limit int, excludeIds []string) ([]*dto.TrackWithDetails, error) {
	var tracks []*dto.TrackWithDetails

	query := m.DB(ctx).
		Table("tracks").
		Select("tracks.*, albums.cover_art as album_cover_url, artists.name as artist_name").
		Joins("JOIN albums ON tracks.album_id = albums.id").
		Joins("JOIN artists ON albums.artist_id = artists.id")

	if len(excludeIds) > 0 {
		query = query.Where("tracks.id NOT IN ?", excludeIds)
	}

	err := query.Order("RANDOM()").Limit(limit).Scan(&tracks).Error
	return tracks, err
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
