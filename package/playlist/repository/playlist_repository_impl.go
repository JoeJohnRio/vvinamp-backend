package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"vvinamp/database/postgre"
	dbModel "vvinamp/model"
	"vvinamp/package/playlist"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrPlaylistNotFound  = errors.New("playlist not found")
	ErrNotOwner          = errors.New("user is not the playlist owner")
	ErrSongAlreadyExists = errors.New("song already exists in playlist")
)

// sqlRepository contains all interactions with the playlist table
type sqlRepository struct {
	logger *zap.Logger
}

// SQLRepositoryTarget is the fx.In struct for dependency injection
type SQLRepositoryTarget struct {
	fx.In
	Connection *postgre.Connection
	Logger     *zap.Logger
}

// NewSQLRepository creates a new playlist repository
func NewSQLRepository(target SQLRepositoryTarget) (playlist.Repository, error) {
	db, err := target.Connection.Client()
	if err != nil {
		return nil, err
	}

	// Auto-migrate the playlist tables
	if err := db.Debug().AutoMigrate(&dbModel.Playlist{}, &dbModel.PlaylistTrack{}, &dbModel.UserSavedPlaylist{}); err != nil {
		return nil, err
	}

	return &sqlRepository{
		logger: target.Logger,
	}, nil
}

// DB extracts the database client from context
func (r *sqlRepository) DB(ctx context.Context) *gorm.DB {
	return postgre.ForContext(ctx)
}

// GetByID retrieves a playlist by its ID
func (r *sqlRepository) GetByID(ctx context.Context, id string) (*dbModel.Playlist, error) {
	var playlist dbModel.Playlist
	err := r.DB(ctx).Where("id = ?", id).First(&playlist).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPlaylistNotFound
		}
		return nil, err
	}
	return &playlist, nil
}

// GetByCreatorID retrieves playlists by creator ID with pagination
func (r *sqlRepository) GetByCreatorID(ctx context.Context, creatorID string, limit int, offset int) ([]*dbModel.Playlist, int64, error) {
	var playlists []*dbModel.Playlist
	var total int64

	// Get total count
	if err := r.DB(ctx).Model(&dbModel.Playlist{}).Where("creator_id = ?", creatorID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.DB(ctx).
		Where("creator_id = ?", creatorID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&playlists).Error

	return playlists, total, err
}

// GetSavedPlaylistsByUserID retrieves playlists saved by a user with pagination
func (r *sqlRepository) GetSavedPlaylistsByUserID(ctx context.Context, userID string, limit int, offset int) ([]*dbModel.Playlist, int64, error) {
	var playlists []*dbModel.Playlist
	var total int64

	// Get total count
	if err := r.DB(ctx).
		Table("user_saved_playlists").
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.DB(ctx).
		Table("playlists").
		Joins("INNER JOIN user_saved_playlists ON playlists.id = user_saved_playlists.playlist_id").
		Where("user_saved_playlists.user_id = ?", userID).
		Order("user_saved_playlists.saved_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&playlists).Error

	return playlists, total, err
}

// GetMyPlaylistFeed retrieves both created and saved playlists for a user, sorted by the specified order
func (r *sqlRepository) GetMyPlaylistFeed(ctx context.Context, userID string, limit int, offset int, sortOrder string) ([]*dbModel.Playlist, int64, error) {
	var playlists []*dbModel.Playlist
	var total int64

	// Build order clause based on sort order
	var orderClause string
	switch sortOrder {
	case "RECENTLY_ADDED":
		// For created playlists: created_at, for saved: saved_at
		// We use a unified approach with COALESCE
		orderClause = "COALESCE(usp.saved_at, p.created_at) DESC"
	case "ALPHABETICAL":
		orderClause = "p.title ASC"
	case "CREATOR":
		orderClause = "u.display_name ASC"
	case "RECENTS", "":
		// For created playlists: updated_at, for saved: saved_at (most recent activity)
		orderClause = "COALESCE(usp.saved_at, p.updated_at, p.created_at) DESC"
	default:
		orderClause = "COALESCE(usp.saved_at, p.updated_at, p.created_at) DESC"
	}

	// Get total count - union of created and saved playlists
	countQuery := `
		SELECT COUNT(DISTINCT p.id)
		FROM playlists p
		LEFT JOIN user_saved_playlists usp ON p.id = usp.playlist_id
		LEFT JOIN users u ON p.creator_id = u.id
		WHERE p.creator_id = ? OR usp.user_id = ?
	`
	if err := r.DB(ctx).Raw(countQuery, userID, userID).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with union of created and saved playlists
	query := `
		SELECT DISTINCT p.*, 
			   COALESCE(usp.saved_at, p.updated_at, p.created_at) AS sort_time
		FROM playlists p
		LEFT JOIN user_saved_playlists usp ON p.id = usp.playlist_id AND usp.user_id = ?
		LEFT JOIN users u ON p.creator_id = u.id
		WHERE p.creator_id = ? OR usp.user_id = ?
		ORDER BY ` + orderClause + `
		LIMIT ? OFFSET ?
	`

	err := r.DB(ctx).Raw(query, userID, userID, userID, limit, offset).Scan(&playlists).Error
	if err != nil {
		return nil, 0, err
	}

	return playlists, total, err
}

// Create creates a new playlist
func (r *sqlRepository) Create(ctx context.Context, playlist *dbModel.Playlist) (*dbModel.Playlist, error) {
	now := time.Now().UTC()
	playlist.CreatedAt = now
	playlist.UpdatedAt = &now

	err := r.DB(ctx).Debug().Create(playlist).Error
	if err != nil {
		return nil, err
	}
	return playlist, nil
}

// Update updates an existing playlist
func (r *sqlRepository) Update(ctx context.Context, playlist *dbModel.Playlist) (*dbModel.Playlist, error) {
	now := time.Now().UTC()
	playlist.UpdatedAt = &now

	err := r.DB(ctx).Debug().Save(playlist).Error
	if err != nil {
		return nil, err
	}
	return playlist, nil
}

// Delete deletes a playlist by ID
func (r *sqlRepository) Delete(ctx context.Context, id string) error {
	return r.DB(ctx).Delete(&dbModel.Playlist{}, id).Error
}

// GetSongsByPlaylistID retrieves all songs in a playlist ordered by position
func (r *sqlRepository) GetSongsByPlaylistID(ctx context.Context, playlistID string) ([]*dbModel.PlaylistTrack, error) {
	var playlistTracks []*dbModel.PlaylistTrack
	err := r.DB(ctx).
		Where("playlist_id = ?", playlistID).
		Order("position ASC").
		Find(&playlistTracks).Error
	return playlistTracks, err
}

// AddSong adds a song to a playlist at the specified position
func (r *sqlRepository) AddSong(ctx context.Context, playlistID, trackID string, position int) error {
	// Check if song already exists in playlist
	var existing dbModel.PlaylistTrack
	err := r.DB(ctx).
		Where("playlist_id = ? AND track_id = ?", playlistID, trackID).
		First(&existing).Error

	if err == nil {
		return ErrSongAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// If position is not specified or negative, add to the end using zero-based indexing.
	if position < 0 {
		var maxPosition int
		err := r.DB(ctx).
			Model(&dbModel.PlaylistTrack{}).
			Where("playlist_id = ?", playlistID).
			Select("COALESCE(MAX(position), -1)").
			Scan(&maxPosition).Error
		if err != nil {
			return err
		}
		position = maxPosition + 1
	}

	playlistTrack := &dbModel.PlaylistTrack{
		PlaylistID: playlistID,
		TrackID:    trackID,
		Position:   position,
		AddedAt:    time.Now().UTC(),
	}

	return r.DB(ctx).Create(playlistTrack).Error
}

// RemoveSong removes a song from a playlist
func (r *sqlRepository) RemoveSong(ctx context.Context, playlistID, trackID string) error {
	return r.DB(ctx).
		Where("playlist_id = ? AND track_id = ?", playlistID, trackID).
		Delete(&dbModel.PlaylistTrack{}).Error
}

// EditPlaylist applies deleted track removals and repositioning in a single transaction.
func (r *sqlRepository) EditPlaylist(ctx context.Context, playlistID string, deletedTrackIds []string, repositionedSongs []playlist.RepositionSong) error {
	tx := r.DB(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if len(deletedTrackIds) > 0 {
		if err := tx.Where("playlist_id = ? AND track_id IN ?", playlistID, deletedTrackIds).
			Delete(&dbModel.PlaylistTrack{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Rebuild the playlist ordering after any deletes and requested moves.
	var playlistTracks []*dbModel.PlaylistTrack
	if err := tx.
		Where("playlist_id = ?", playlistID).
		Order("position ASC").
		Find(&playlistTracks).Error; err != nil {
		tx.Rollback()
		return err
	}

	trackOrder := make([]string, 0, len(playlistTracks))
	for _, pt := range playlistTracks {
		trackOrder = append(trackOrder, pt.TrackID)
	}

	for _, song := range repositionedSongs {
		currentIndex := -1
		for idx, trackID := range trackOrder {
			if trackID == song.TrackID {
				currentIndex = idx
				break
			}
		}
		if currentIndex < 0 {
			tx.Rollback()
			return fmt.Errorf("playlist track not found: %s", song.TrackID)
		}

		trackOrder = append(trackOrder[:currentIndex], trackOrder[currentIndex+1:]...)

		target := song.Position
		if target < 0 {
			target = 0
		}
		if target > len(trackOrder) {
			target = len(trackOrder)
		}

		trackOrder = append(trackOrder[:target], append([]string{song.TrackID}, trackOrder[target:]...)...)
	}

	// Update all track positions sequentially using zero-based indexing.
	for idx, trackID := range trackOrder {
		newPos := idx
		if err := tx.Model(&dbModel.PlaylistTrack{}).
			Where("playlist_id = ? AND track_id = ?", playlistID, trackID).
			Update("position", newPos).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// Save saves a playlist for a user
func (r *sqlRepository) Save(ctx context.Context, userID, playlistID string) error {
	// Check if already saved
	var existing dbModel.UserSavedPlaylist
	err := r.DB(ctx).
		Where("user_id = ? AND playlist_id = ?", userID, playlistID).
		First(&existing).Error

	if err == nil {
		// Already saved, return nil (idempotent)
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	userSavedPlaylist := &dbModel.UserSavedPlaylist{
		UserID:     userID,
		PlaylistID: playlistID,
		SavedAt:    time.Now().UTC(),
	}

	return r.DB(ctx).Create(userSavedPlaylist).Error
}

// Unsave removes a playlist from a user's saved list
func (r *sqlRepository) Unsave(ctx context.Context, userID, playlistID string) error {
	result := r.DB(ctx).
		Where("user_id = ? AND playlist_id = ?", userID, playlistID).
		Delete(&dbModel.UserSavedPlaylist{})

	if result.RowsAffected == 0 {
		return nil // Already unsaved, return nil (idempotent)
	}

	return result.Error
}

// IsSavedByUser checks if a playlist is saved by a user
func (r *sqlRepository) IsSavedByUser(ctx context.Context, userID, playlistID string) (bool, error) {
	var count int64
	err := r.DB(ctx).
		Model(&dbModel.UserSavedPlaylist{}).
		Where("user_id = ? AND playlist_id = ?", userID, playlistID).
		Count(&count).Error
	return count > 0, err
}

// GetSavedCount returns the number of users who saved a playlist
func (r *sqlRepository) GetSavedCount(ctx context.Context, playlistID string) (int64, error) {
	var count int64
	err := r.DB(ctx).
		Model(&dbModel.UserSavedPlaylist{}).
		Where("playlist_id = ?", playlistID).
		Count(&count).Error
	return count, err
}

// GetTotalDuration returns the total duration of all tracks in a playlist
func (r *sqlRepository) GetTotalDuration(ctx context.Context, playlistID string) (int, error) {
	var total int
	err := r.DB(ctx).
		Table("tracks").
		Joins("INNER JOIN playlist_track ON tracks.id = playlist_track.track_id").
		Where("playlist_track.playlist_id = ?", playlistID).
		Select("COALESCE(SUM(tracks.duration), 0)").
		Scan(&total).Error
	return total, err
}

// IsOwner checks if a user is the owner of a playlist
func (r *sqlRepository) IsOwner(ctx context.Context, playlistID, userID string) (bool, error) {
	var count int64
	err := r.DB(ctx).
		Model(&dbModel.Playlist{}).
		Where("id = ? AND creator_id = ?", playlistID, userID).
		Count(&count).Error
	return count > 0, err
}

// Ensure the repository implements the interface
var _ playlist.Repository = (*sqlRepository)(nil)
