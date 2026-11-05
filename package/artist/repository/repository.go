package repository

import (
	"context"
	"errors"

	"vvinamp/database/postgre"
	dbModel "vvinamp/model"
	"vvinamp/model/dto"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type ArtistRepository interface {
	GetArtistByID(ctx context.Context, id string, userID *string) (*dbModel.Artist, error)
	IsUserFollowingArtist(ctx context.Context, userID, artistID string) (bool, error)
	GetTopSongsByArtist(ctx context.Context, artistID string, userID *string) ([]*dto.TrackWithDetails, error)
	GetTopReleasesByArtist(ctx context.Context, artistID string, userID *string) ([]*dbModel.Album, error)
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

	if err := db.Debug().AutoMigrate(
		&dbModel.Artist{},
		&dbModel.PodcastSeries{},
		&dbModel.PodcastEpisode{},
		&dbModel.Playlist{},
	); err != nil {
		return nil, err
	}

	return &sqlRepository{}, nil
}

// DB method extract database client from context
func (m *sqlRepository) DB(ctx context.Context) *gorm.DB {
	return postgre.ForContext(ctx)
}

func (r *sqlRepository) GetArtistByID(ctx context.Context, id string, userID *string) (*dbModel.Artist, error) {
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

func (r *sqlRepository) GetTopSongsByArtist(ctx context.Context, artistID string, userID *string) ([]*dto.TrackWithDetails, error) {
	var tracks []*dto.TrackWithDetails // Changed from []*dbModel.Track

	// Find tracks by joining through albums (tracks -> albums -> artists)
	// Also join to get album cover and artist name
	err := r.DB(ctx).
		Table("tracks").
		Select("tracks.*, albums.cover_art as album_cover_url, artists.name as artist_name, artists.bio as artist_bio, artists.profile_image as artist_profile_image_url, artists.monthly_listeners as artist_monthly_listeners").
		Joins("JOIN albums ON tracks.album_id = albums.id").
		Joins("JOIN artists ON albums.artist_id = artists.id"). // Added to get artist name
		Where("albums.artist_id::text = ?", artistID).
		Order("tracks.stream_count DESC").
		Limit(10).
		Scan(&tracks).Error // Changed from Find to Scan

	if err != nil {
		return nil, err
	}

	// If no tracks found through albums, try the track_artists junction table as fallback
	if len(tracks) == 0 {
		err = r.DB(ctx).
			Table("tracks").
			Select("tracks.*, albums.cover_art as album_cover_url, artists.name as artist_name").
			Joins("JOIN track_artists ON tracks.id = track_artists.track_id").
			Joins("JOIN artists ON track_artists.artist_id = artists.id"). // Get artist name
			Joins("JOIN albums ON tracks.album_id = albums.id").           // Get album cover
			Where("track_artists.artist_id = ?", artistID).
			Order("tracks.stream_count DESC").
			Limit(10).
			Scan(&tracks).Error // Changed from Find to Scan

		if err != nil {
			return nil, err
		}
	}

	return tracks, nil
}

func (r *sqlRepository) GetTopReleasesByArtist(ctx context.Context, artistID string, userID *string) ([]*dbModel.Album, error) {
	var albums []*dbModel.Album

	// Get albums by the artist ordered by popularity (sum of stream counts of tracks in each album)
	// This complex query joins albums with tracks to calculate popularity based on total stream count
	err := r.DB(ctx).
		Table("albums").
		Joins("LEFT JOIN tracks ON albums.id = tracks.album_id").
		Where("albums.artist_id::text = ?", artistID).
		Group("albums.id").
		Order("COALESCE(SUM(tracks.stream_count), 0) DESC, albums.release_date DESC"). // Order by total stream count, then by release date as fallback
		Limit(10).
		Find(&albums).Error

	if err != nil {
		return nil, err
	}

	return albums, nil
}
