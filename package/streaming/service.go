package streaming

import (
	"context"
	"fmt"
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
}

// NewService creates a new streaming service
func NewService(target ServiceTarget) *Service {
	redisClient, _ := target.RedisConn.Client()
	return &Service{
		FileController: target.FileController,
		Redis:          redisClient,
	}
}

// GetPlaybackInfo returns the HLS URL for a given track
func (s *Service) GetPlaybackInfo(ctx context.Context, sourcePath string, contentType string, contentId string) (*PlaybackInfoResult, error) {

	trackIdUuid, err := uuid.Parse(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("invalid track id: %w", err)
	}

	// 1. GET THE REAL PATH using the new helper
	// The Service asks the FileController: "Where is the MP3 for this track?"
	mp3Path, err := s.FileController.GetTrackPath(ctx, trackIdUuid)
	if err != nil {
		return nil, fmt.Errorf("could not find mp3 path: %w", err)
	}

	if !s.FileController.HLSExists(sourcePath) {
		if _, err := s.FileController.ConvertToHLS(ctx, mp3Path, sourcePath); err != nil {
			return nil, fmt.Errorf("hls convert: %w", err)
		}
	}

	// --- AUTO-RECORD HISTORY & UPDATE COUNTER ---
	userId, _ := ctx.Value(constant.Session).(string)
	if userId != "" {
		db := postgre.ForContext(ctx)

		actualContentType := "TRACK"
		if contentType != "" {
			actualContentType = contentType
		}

		actualContentId := sourcePath
		if contentId != "" {
			actualContentId = contentId
		}

		// 1. Remove old history entry for this content (to move it to the top)
		db.Where("user_id = ? AND content_id = ? AND content_type = ?", userId, actualContentId, actualContentType).
			Delete(&model.ListeningHistory{})

		// 2. Create new history entry
		historyEntry := model.ListeningHistory{
			UserID:      userId,
			ContentID:   actualContentId,
			ContentType: actualContentType,
			PlayedAt:    time.Now(),
		}
		if err := db.Create(&historyEntry).Error; err != nil {
			fmt.Printf("failed to record history: %v\n", err)
		}

		// 3. Increment the Track's stream count
		db.Model(&model.Track{}).
			Where("id = ?", sourcePath).
			Update("stream_count", gorm.Expr("stream_count + ?", 1))
	}
	// -------------------------------------------
	// ---------------------------

	// Note: The actual base URL should come from configuration
	// For now, returning just the path part which will be combined with the server's base URL
	return &PlaybackInfoResult{
		HlsURL: fmt.Sprintf("/hls/%s/playlist.m3u8", sourcePath),
		// Mp3Fallback: fmt.Sprintf("/tracks/%s.mp3", trackID), // optional fallback
	}, nil
}

// PlaybackInfoResult represents the result of a playback info request
type PlaybackInfoResult struct {
	HlsURL      string
	Mp3Fallback string
	DurationMs  *int
}
