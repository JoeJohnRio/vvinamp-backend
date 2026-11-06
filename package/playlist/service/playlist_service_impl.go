package service

import (
	"context"
	"fmt"

	"vvinamp/database/postgre"
	dbModel "vvinamp/model"
	"vvinamp/package/playlist"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// service implements the playlist.Service interface
type service struct {
	repo   playlist.Repository
	logger *zap.Logger
}

// ServiceTarget is the fx.In struct for dependency injection
type ServiceTarget struct {
	fx.In
	Repository playlist.Repository
	Logger     *zap.Logger
	Connection *postgre.Connection
}

// NewService creates a new playlist service
func NewService(target ServiceTarget) playlist.Service {
	return &service{
		repo:   target.Repository,
		logger: target.Logger,
	}
}

// GetPlaylistByID retrieves a playlist by ID
func (s *service) GetPlaylistByID(ctx context.Context, id string) (*dbModel.Playlist, error) {
	return s.repo.GetByID(ctx, id)
}

// GetPlaylistsByCreatorID retrieves playlists by creator ID with pagination
func (s *service) GetPlaylistsByCreatorID(ctx context.Context, creatorID string, limit int, offset int) ([]*dbModel.Playlist, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	return s.repo.GetByCreatorID(ctx, creatorID, limit, offset)
}

// GetMyPlaylists retrieves playlists created by the user
func (s *service) GetMyPlaylists(ctx context.Context, userID string, limit int, offset int) ([]*dbModel.Playlist, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	return s.repo.GetByCreatorID(ctx, userID, limit, offset)
}

// GetMyPlaylistFeed retrieves both created and saved playlists for a user with sorting
func (s *service) GetMyPlaylistFeed(ctx context.Context, userID string, limit int, offset int, sortOrder string) ([]*dbModel.Playlist, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	return s.repo.GetMyPlaylistFeed(ctx, userID, limit, offset, sortOrder)
}

// CreatePlaylist creates a new playlist
func (s *service) CreatePlaylist(ctx context.Context, playlist *dbModel.Playlist) (*dbModel.Playlist, error) {
	return s.repo.Create(ctx, playlist)
}

// UpdatePlaylist updates an existing playlist
func (s *service) UpdatePlaylist(ctx context.Context, playlist *dbModel.Playlist) (*dbModel.Playlist, error) {
	return s.repo.Update(ctx, playlist)
}

// DeletePlaylist deletes a playlist
func (s *service) DeletePlaylist(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// AddSongToPlaylist adds a song to a playlist
func (s *service) AddSongToPlaylist(ctx context.Context, playlistID, trackID string, position int) error {
	return s.repo.AddSong(ctx, playlistID, trackID, position)
}

// AddYoutubeTrackToPlaylist adds a YouTube track to a playlist
func (s *service) AddYoutubeTrackToPlaylist(ctx context.Context, playlistID, youtubeID, title, artistName, thumbnailURL string, duration int, position int) error {
	return s.repo.AddYoutubeTrack(ctx, playlistID, youtubeID, title, artistName, thumbnailURL, duration, position)
}

// RemoveSongFromPlaylist removes a song from a playlist
func (s *service) RemoveSongFromPlaylist(ctx context.Context, playlistID, trackID string) error {
	return s.repo.RemoveSong(ctx, playlistID, trackID)
}

// EditPlaylist applies deletes and repositioning to a playlist in a single operation.
func (s *service) EditPlaylist(ctx context.Context, playlistID string, deletedTrackIds []string, repositionedSongs []playlist.RepositionSong) error {
	return s.repo.EditPlaylist(ctx, playlistID, deletedTrackIds, repositionedSongs)
}

// GetPlaylistSongs retrieves all songs in a playlist
func (s *service) GetPlaylistSongs(ctx context.Context, playlistID string) ([]*dbModel.PlaylistTrack, error) {
	return s.repo.GetSongsByPlaylistID(ctx, playlistID)
}

// SavePlaylist saves a playlist for a user
func (s *service) SavePlaylist(ctx context.Context, userID, playlistID string) error {
	// Verify playlist exists
	_, err := s.repo.GetByID(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("playlist not found: %w", err)
	}
	return s.repo.Save(ctx, userID, playlistID)
}

// UnsavePlaylist removes a playlist from a user's saved list
func (s *service) UnsavePlaylist(ctx context.Context, userID, playlistID string) error {
	return s.repo.Unsave(ctx, userID, playlistID)
}

// IsPlaylistSavedByUser checks if a playlist is saved by a user
func (s *service) IsPlaylistSavedByUser(ctx context.Context, userID, playlistID string) (bool, error) {
	return s.repo.IsSavedByUser(ctx, userID, playlistID)
}

// GetPlaylistSavedCount returns the number of users who saved a playlist
func (s *service) GetPlaylistSavedCount(ctx context.Context, playlistID string) (int64, error) {
	return s.repo.GetSavedCount(ctx, playlistID)
}

// GetPlaylistTotalDuration returns the total duration of all tracks in a playlist
func (s *service) GetPlaylistTotalDuration(ctx context.Context, playlistID string) (int, error) {
	return s.repo.GetTotalDuration(ctx, playlistID)
}

// IsPlaylistOwner checks if a user is the owner of a playlist
func (s *service) IsPlaylistOwner(ctx context.Context, playlistID, userID string) (bool, error) {
	return s.repo.IsOwner(ctx, playlistID, userID)
}

// Ensure the service implements the interface
var _ playlist.Service = (*service)(nil)
