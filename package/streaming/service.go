package streaming

import (
	"context"
	"fmt"
	"strings"
	"time"
	"vvinamp/constant"
	"vvinamp/database/postgre"
	redisDb "vvinamp/database/redis"
	"vvinamp/model"
	"vvinamp/package/song"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

// HLSConverter interface defines the methods needed for HLS conversion
type HLSConverter interface {
	ConvertToHLS(ctx context.Context, sourcePath, trackId string) (string, error)
	GetTrackPath(ctx context.Context, trackID uuid.UUID) (string, error)
	HLSExists(trackId string) bool
}

// Service handles streaming-related operations
type Service struct {
	Repo           song.Repository
	FileController HLSConverter
	Redis          *redis.Client
}

// ServiceTarget is parameter object for getting all Service's dependencies
type ServiceTarget struct {
	fx.In
	FileController HLSConverter `name:"hlsConverter"`
	RedisConn      *redisDb.Connection
	Repo           song.Repository // Add this line!
}

// NewService creates a new streaming service
func NewService(target ServiceTarget) *Service {
	redisClient, _ := target.RedisConn.Client()
	return &Service{
		FileController: target.FileController,
		Redis:          redisClient,
		Repo:           target.Repo, // Add this line!
	}
}

// GetPlaybackInfo returns the HLS or cloud audio URL for a given track
func (s *Service) GetPlaybackInfo(ctx context.Context, sourcePath string, contentType string, contentId string) (*PlaybackInfoResult, error) {
	trackIdUuid, err := uuid.Parse(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("invalid track id: %w", err)
	}

	// 1. If the track has a cloud audio URL in the database (e.g. Supabase Storage CDN), serve it directly
	track, err := s.Repo.GetTrackByID(ctx, sourcePath)
	if err == nil && track != nil && track.AudioFileURL != "" && strings.HasPrefix(track.AudioFileURL, "http") {
		s.recordListeningHistory(ctx, sourcePath, contentType, contentId)
		return &PlaybackInfoResult{
			HlsURL:      track.AudioFileURL,
			Mp3Fallback: track.AudioFileURL,
		}, nil
	}

	// 2. Fall back to local HLS conversion if running locally with disk files
	mp3Path, err := s.FileController.GetTrackPath(ctx, trackIdUuid)
	if err == nil && !s.FileController.HLSExists(sourcePath) {
		_, _ = s.FileController.ConvertToHLS(ctx, mp3Path, sourcePath)
	}

	s.recordListeningHistory(ctx, sourcePath, contentType, contentId)

	return &PlaybackInfoResult{
		HlsURL: fmt.Sprintf("/hls/%s/playlist.m3u8", sourcePath),
	}, nil
}

func (s *Service) recordListeningHistory(ctx context.Context, sourcePath, contentType, contentId string) {
	userId, _ := ctx.Value(constant.Session).(string)
	if userId == "" {
		return
	}
	db := postgre.ForContext(ctx)

	actualContentType := "TRACK"
	if contentType != "" {
		actualContentType = contentType
	}

	actualContentId := sourcePath
	if contentId != "" {
		actualContentId = contentId
	}

	// Remove old history entry for this content (to move it to the top)
	db.Where("user_id = ? AND content_id = ? AND content_type = ?", userId, actualContentId, actualContentType).
		Delete(&model.ListeningHistory{})

	// Create new history entry
	historyEntry := model.ListeningHistory{
		UserID:      userId,
		ContentID:   actualContentId,
		ContentType: actualContentType,
		PlayedAt:    time.Now(),
	}
	if err := db.Create(&historyEntry).Error; err != nil {
		fmt.Printf("failed to record history: %v\n", err)
	}

	// Increment the Track's stream count
	db.Model(&model.Track{}).
		Where("id = ?", sourcePath).
		Update("stream_count", gorm.Expr("stream_count + ?", 1))
}

// PlaybackInfoResult represents the result of a playback info request
type PlaybackInfoResult struct {
	HlsURL      string
	Mp3Fallback string
	DurationMs  *int
}
