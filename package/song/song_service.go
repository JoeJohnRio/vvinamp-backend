package song

import (
	"context"

	dbModel "vvinamp/model"
	"vvinamp/model/dto"
)

type key string

const (
	Key key = "song"
)

// Service is interface for UserService
type Service interface {
	Repository
	GetSingleTrack(ctx context.Context, trackId string) (*dbModel.Track, error)
	GetAlbumByID(ctx context.Context, albumId string) (*dbModel.Album, error)
	GetTracksByAlbumID(ctx context.Context, albumId string) ([]*dbModel.Track, error)
	GetTracksByAlbumIDWithDetails(ctx context.Context, albumId string) ([]*dto.TrackWithDetails, error)
}

// ForContext is method to get user service from context
func ForContext(ctx context.Context) Service {
	service, ok := ctx.Value(Key).(Service)
	if !ok {
		panic("ctx passing is not contain song service")
	}
	return service
}
