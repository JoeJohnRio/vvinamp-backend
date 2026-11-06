package playlist

import (
	"context"

	dbModel "vvinamp/model"
)

// Repository defines the interface for playlist data access
type Repository interface {
	// Basic CRUD
	GetByID(ctx context.Context, id string) (*dbModel.Playlist, error)
	GetByCreatorID(ctx context.Context, creatorID string, limit int, offset int) ([]*dbModel.Playlist, int64, error)
	Create(ctx context.Context, playlist *dbModel.Playlist) (*dbModel.Playlist, error)
	Update(ctx context.Context, playlist *dbModel.Playlist) (*dbModel.Playlist, error)
	Delete(ctx context.Context, id string) error

	// Playlist songs
	GetSongsByPlaylistID(ctx context.Context, playlistID string) ([]*dbModel.PlaylistTrack, error)
	AddSong(ctx context.Context, playlistID, trackID string, position int) error
	AddYoutubeTrack(ctx context.Context, playlistID, youtubeID, title, artistName, thumbnailURL string, duration int, position int) error
	RemoveSong(ctx context.Context, playlistID, trackID string) error
	EditPlaylist(ctx context.Context, playlistID string, deletedTrackIds []string, repositionedSongs []RepositionSong) error

	// Saved playlists
	Save(ctx context.Context, userID, playlistID string) error
	Unsave(ctx context.Context, userID, playlistID string) error
	IsSavedByUser(ctx context.Context, userID, playlistID string) (bool, error)
	GetSavedCount(ctx context.Context, playlistID string) (int64, error)
	GetSavedPlaylistsByUserID(ctx context.Context, userID string, limit int, offset int) ([]*dbModel.Playlist, int64, error)
	GetMyPlaylistFeed(ctx context.Context, userID string, limit int, offset int, sortOrder string) ([]*dbModel.Playlist, int64, error)

	// Stats
	GetTotalDuration(ctx context.Context, playlistID string) (int, error)

	// Authorization
	IsOwner(ctx context.Context, playlistID, userID string) (bool, error)
}
