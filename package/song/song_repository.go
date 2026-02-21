package song

import (
	"context"
	dbModel "vvinamp/model"
	"vvinamp/model/dto"
)

type Repository interface {
	GetTrackByID(ctx context.Context, trackId string) (*dbModel.Track, error)
	GetAlbumByID(ctx context.Context, albumId string) (*dbModel.Album, error)
	GetTracksByAlbumID(ctx context.Context, albumId string) ([]*dbModel.Track, error)
	GetTracksByAlbumIDWithDetails(ctx context.Context, albumId string) ([]*dto.TrackWithDetails, error)

	// ... other methods as needed
}
