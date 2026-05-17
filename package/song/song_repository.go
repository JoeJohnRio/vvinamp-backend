package song

import (
	"context"
	dbModel "vvinamp/model"
	"vvinamp/model/dto"
)

type Repository interface {
	GetTracksWithDetails(ctx context.Context, trackIds []string) ([]*dto.TrackWithDetails, error)
	GetTrackByID(ctx context.Context, trackId string) (*dbModel.Track, error)
	GetAlbumByID(ctx context.Context, albumId string) (*dbModel.Album, error)
	GetTracksByAlbumID(ctx context.Context, albumId string) ([]*dbModel.Track, error)
	GetTracksByAlbumIDWithDetails(ctx context.Context, albumId string) ([]*dto.TrackWithDetails, error)
	GetTracksByGenres(ctx context.Context, genreIds []string, limit int, excludeIds []string) ([]*dto.TrackWithDetails, error)
	GetRandomTracks(ctx context.Context, limit int, excludeIds []string) ([]*dto.TrackWithDetails, error)
}
