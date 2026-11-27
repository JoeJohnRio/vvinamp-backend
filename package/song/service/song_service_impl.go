package song

import (
	"context"
	"vvinamp/constant"
	"vvinamp/database/postgre"
	"vvinamp/model"
	"vvinamp/model/dto"
	"vvinamp/package/song"

	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

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
// It picks random songs from the listener's three most-played genres, leaves out songs they played recently,
// and tops the list up from the whole catalogue when those genres run short.
func (s *songService) GetRecommendedTracks(ctx context.Context, limit int) ([]*dto.TrackWithDetails, error) {
	db := postgre.ForContext(ctx)
	userId, _ := ctx.Value(constant.Session).(string)

	var topGenres, played []string
	if userId != "" {
		db.Table("listening_history").
			Select("albums.genre_id").
			Joins("JOIN tracks ON listening_history.content_id = tracks.id::text").
			Joins("JOIN albums ON tracks.album_id = albums.id").
			Where("listening_history.user_id = ? AND listening_history.content_type = ? AND albums.genre_id IS NOT NULL", userId, "TRACK").
			Group("albums.genre_id").
			Order("count(*) DESC").
			Limit(3).
			Pluck("genre_id", &topGenres)

		var recent []string
		db.Table("listening_history").
			Where("user_id = ? AND content_type = ?", userId, "TRACK").
			Order("played_at DESC").
			Limit(50).
			Pluck("content_id", &recent)
		played = trackIDsOnly(recent)
	}

	tracks, err := s.Repository.GetTracksByGenres(ctx, topGenres, limit, played)
	if err != nil {
		tracks = nil
	}
	if len(tracks) < limit {
		exclude := played
		for _, t := range tracks {
			exclude = append(exclude, t.ID.String())
		}
		if more, err := s.Repository.GetRandomTracks(ctx, limit-len(tracks), exclude); err == nil {
			tracks = append(tracks, more...)
		}
	}
	if len(tracks) == 0 {
		// Everything was played recently: show anything rather than an empty Home.
		return s.Repository.GetRandomTracks(ctx, limit, nil)
	}
	return tracks, nil
}

// trackIDsOnly drops history IDs that aren't song IDs (an unsaved YouTube video ID, say), which the database
// would reject when compared with the tracks' UUID column.
func trackIDsOnly(ids []string) []string {
	valid := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, err := uuid.Parse(id); err == nil {
			valid = append(valid, id)
		}
	}
	return valid
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
