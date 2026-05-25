package playlist

import (
	"context"

	dbModel "vvinamp/model"
)

type key string

type RepositionSong struct {
	TrackID  string
	Position int
}

const (
	Key key = "playlist"
)

// Service is the interface for PlaylistService
type Service interface {
	GetPlaylistByID(ctx context.Context, id string) (*dbModel.Playlist, error)
	GetPlaylistsByCreatorID(ctx context.Context, creatorID string, limit int, offset int) ([]*dbModel.Playlist, int64, error)
	GetMyPlaylists(ctx context.Context, userID string, limit int, offset int) ([]*dbModel.Playlist, int64, error)
	GetMyPlaylistFeed(ctx context.Context, userID string, limit int, offset int, sortOrder string) ([]*dbModel.Playlist, int64, error)
	CreatePlaylist(ctx context.Context, playlist *dbModel.Playlist) (*dbModel.Playlist, error)
	UpdatePlaylist(ctx context.Context, playlist *dbModel.Playlist) (*dbModel.Playlist, error)
	DeletePlaylist(ctx context.Context, id string) error
	AddSongToPlaylist(ctx context.Context, playlistID, trackID string, position int) error
	RemoveSongFromPlaylist(ctx context.Context, playlistID, trackID string) error
	EditPlaylist(ctx context.Context, playlistID string, deletedTrackIds []string, repositionedSongs []RepositionSong) error
	GetPlaylistSongs(ctx context.Context, playlistID string) ([]*dbModel.PlaylistTrack, error)
	SavePlaylist(ctx context.Context, userID, playlistID string) error
	UnsavePlaylist(ctx context.Context, userID, playlistID string) error
	IsPlaylistSavedByUser(ctx context.Context, userID, playlistID string) (bool, error)
	GetPlaylistSavedCount(ctx context.Context, playlistID string) (int64, error)
	GetPlaylistTotalDuration(ctx context.Context, playlistID string) (int, error)
	IsPlaylistOwner(ctx context.Context, playlistID, userID string) (bool, error)
}

// ForContext is the method to get playlist service from context
func ForContext(ctx context.Context) Service {
	service, ok := ctx.Value(Key).(Service)
	if !ok {
		panic("ctx passing does not contain playlist service")
	}
	return service
}
