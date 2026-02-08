package streaming

import (
	"context"
	"fmt"

	"go.uber.org/fx"
)

// HLSConverter interface defines the methods needed for HLS conversion
type HLSConverter interface {
	ConvertToHLS(ctx context.Context, mp3RelativePath, trackId string) (string, error)
	HLSExists(trackId string) bool
}

// Service handles streaming-related operations
type Service struct {
	FileController HLSConverter
}

// ServiceTarget is parameter object for getting all Service's dependencies
type ServiceTarget struct {
	fx.In
	FileController HLSConverter `name:"hlsConverter"`
}

// NewService creates a new streaming service
func NewService(target ServiceTarget) *Service {
	return &Service{
		FileController: target.FileController,
	}
}

// GetPlaybackInfo returns the HLS URL for a given track
func (s *Service) GetPlaybackInfo(ctx context.Context, trackID string) (*PlaybackInfoResult, error) {
	mp3Path := fmt.Sprintf("tracks/%s.mp3", trackID) // adjust to your file layout

	if !s.FileController.HLSExists(trackID) {
		if _, err := s.FileController.ConvertToHLS(ctx, mp3Path, trackID); err != nil {
			return nil, fmt.Errorf("hls convert: %w", err)
		}
	}

	// Note: The actual base URL should come from configuration
	// For now, returning just the path part which will be combined with the server's base URL
	return &PlaybackInfoResult{
		HlsURL: fmt.Sprintf("/hls/%s/playlist.m3u8", trackID),
		// Mp3Fallback: fmt.Sprintf("/tracks/%s.mp3", trackID), // optional fallback
	}, nil
}

// PlaybackInfoResult represents the result of a playback info request
type PlaybackInfoResult struct {
	HlsURL      string
	Mp3Fallback string
	DurationMs  *int
}