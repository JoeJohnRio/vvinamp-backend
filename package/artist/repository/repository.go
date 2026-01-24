package repository

import (
	"context"
	"errors"

	"vvinamp/database/postgre"
	dbModel "vvinamp/model"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type ArtistRepository interface {
	GetArtistHeaderByID(ctx context.Context, id string) (*dbModel.Artist, error)
	IsUserFollowingArtist(ctx context.Context, userID, artistID string) (bool, error)
	GetTopSongsByArtist(ctx context.Context, artistID string) ([]*dbModel.Track, error)
	GetTopReleasesByArtist(ctx context.Context, artistID string) ([]*dbModel.Album, error)
}

// sqlRepository contains all the interactions with the artist table stored in sql.
type sqlRepository struct {
}

// SQLRepositoryTarget is `fx.In` struct for `fx` to get all dependency to create `ArtistSQLRepository`
type SQLRepositoryTarget struct {
	fx.In
	Connection *postgre.Connection
}

func NewSQLRepository(target SQLRepositoryTarget) (ArtistRepository, error) {
	db, err := target.Connection.Client()
	if err != nil {
		return nil, err
	}

	if err := db.Debug().AutoMigrate(&dbModel.Artist{}); err != nil {
		return nil, err
	}

	return &sqlRepository{}, nil
}

// DB method extract database client from context
func (m *sqlRepository) DB(ctx context.Context) *gorm.DB {
	return postgre.ForContext(ctx)
}

func (r *sqlRepository) GetArtistHeaderByID(ctx context.Context, id string) (*dbModel.Artist, error) {
	var artist dbModel.Artist

	if err := r.DB(ctx).Where("id = ?", id).First(&artist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("artist not found")
		}
		return nil, err
	}

	return &artist, nil
}

func (r *sqlRepository) IsUserFollowingArtist(ctx context.Context, userID, artistID string) (bool, error) {
	var count int64

	// Check if user_follows table exists
	if err := r.DB(ctx).Migrator().HasTable("user_follow"); err == false {
		// If table doesn't exist, return false (not following)
		return false, nil
	}

	err := r.DB(ctx).
		Table("user_follow").
		Where("user_id = ? AND artist_id = ?", userID, artistID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *sqlRepository) GetTopSongsByArtist(ctx context.Context, artistID string) ([]*dbModel.Track, error) {
	var tracks []*dbModel.Track

	// Join tracks with track_artists to find tracks by the specific artist
	err := r.DB(ctx).
		Table("tracks").
		Joins("JOIN track_artists ON tracks.id = track_artists.track_id").
		Where("track_artists.artist_id = ?", artistID).
		Order("tracks.stream_count DESC").
		Limit(10).
		Find(&tracks).Error

	if err != nil {
		return nil, err
	}

	return tracks, nil
}

func (r *sqlRepository) GetTopReleasesByArtist(ctx context.Context, artistID string) ([]*dbModel.Album, error) {
	var albums []*dbModel.Album

	// Get albums by the artist ordered by popularity (sum of stream counts of tracks in each album)
	// This complex query joins albums with tracks to calculate popularity based on total stream count
	err := r.DB(ctx).
		Table("albums").
		Joins("LEFT JOIN tracks ON albums.id = tracks.album_id").
		Where("albums.artist_id = ?", artistID).
		Group("albums.id").
		Order("COALESCE(SUM(tracks.stream_count), 0) DESC, albums.release_date DESC"). // Order by total stream count, then by release date as fallback
		Limit(10).
		Find(&albums).Error

	if err != nil {
		return nil, err
	}

	return albums, nil
}
