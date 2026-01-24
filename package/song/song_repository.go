package song

import (
	"context"
	dbModel "vvinamp/model"
)

type Repository interface {
	GetTrackByID(ctx context.Context, trackId string) (*dbModel.Track, error)
	GetAlbumByID(ctx context.Context, albumId string) (*dbModel.Album, error)

	// ... other methods as needed
}
