package song

import (
	"context"
	"vvinamp/model"
	"vvinamp/model/dto"
	"vvinamp/package/song"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// delete check if its used joel123
type Service struct {
	Repo song.Repository
}

type songService struct {
	song.Repository
	logger *zap.Logger
}

// GetTrackById implements song.Service.
func (s *songService) GetSingleTrack(ctx context.Context, trackId string) (*model.Track, error) {
	return s.Repository.GetTrackByID(ctx, trackId)
}

// GetAlbumByID implements song.Service.
func (s *songService) GetAlbumByID(ctx context.Context, albumId string) (*model.Album, error) {
	return s.Repository.GetAlbumByID(ctx, albumId)
}

// GetTracksByAlbumID implements song.Service.
func (s *songService) GetTracksByAlbumID(ctx context.Context, albumId string) ([]*model.Track, error) {
	return s.Repository.GetTracksByAlbumID(ctx, albumId)
}

// GetTracksByAlbumIDWithDetails implements song.Service.
func (s *songService) GetTracksByAlbumIDWithDetails(ctx context.Context, albumId string) ([]*dto.TrackWithDetails, error) {
	return s.Repository.GetTracksByAlbumIDWithDetails(ctx, albumId)
}

type Target struct {
	fx.In
	Repository song.Repository
	Logger     *zap.Logger
}

func NewService(target Target) song.Service {
	return &songService{
		Repository: target.Repository,
		logger:     target.Logger,
	}
}
