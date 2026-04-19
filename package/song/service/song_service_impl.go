package song

import (
	"context"
	"vvinamp/constant"
	"vvinamp/database/postgre"
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

// GetRecommendedTracks implements song.Service.
func (s *songService) GetRecommendedTracks(ctx context.Context, limit int) ([]*dto.TrackWithDetails, error) {
	db := postgre.ForContext(ctx)
	userId, _ := ctx.Value(constant.Session).(string)

	var topGenres []string
	if userId != "" {
		// Get top genres from user history
		db.Table("listening_history").
			Select("albums.genre_id").
			Joins("JOIN tracks ON listening_history.content_id = tracks.id::text").
			Joins("JOIN albums ON tracks.album_id = albums.id").
			Where("listening_history.user_id = ? AND listening_history.content_type = ?", userId, "TRACK").
			Group("albums.genre_id").
			Order("count(*) DESC").
			Limit(3).
			Pluck("genre_id", &topGenres)
	}

	tracks, err := s.Repository.GetTracksByGenres(ctx, topGenres, limit, nil)
	if err != nil || len(tracks) == 0 {
		// Fallback to random if no history or error
		return s.Repository.GetRandomTracks(ctx, limit, nil)
	}

	return tracks, nil
}

// GetRandomTracks implements song.Service.
func (s *songService) GetRandomTracks(ctx context.Context, limit int, excludeIds []string) ([]*dto.TrackWithDetails, error) {
	return s.Repository.GetRandomTracks(ctx, limit, excludeIds)
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
