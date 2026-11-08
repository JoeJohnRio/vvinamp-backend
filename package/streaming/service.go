package streaming

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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

type cachedStream struct {
	url       string
	expiresAt time.Time
}

// Service handles streaming-related operations
type Service struct {
	Repo           song.Repository
	FileController HLSConverter
	Redis          *redis.Client
	memoryCache    sync.Map
}

// ServiceTarget is parameter object for getting all Service's dependencies
type ServiceTarget struct {
	fx.In
	FileController HLSConverter `name:"hlsConverter"`
	RedisConn      *redisDb.Connection
	Repo           song.Repository
}

// NewService creates a new streaming service
func NewService(target ServiceTarget) *Service {
	redisClient, _ := target.RedisConn.Client()
	return &Service{
		FileController: target.FileController,
		Redis:          redisClient,
		Repo:           target.Repo,
	}
}

// GetPlaybackInfo returns the playable audio URL (Supabase CDN, live on-demand YouTube stream, or local HLS)
func (s *Service) GetPlaybackInfo(ctx context.Context, sourcePath string, contentType string, contentId string) (*PlaybackInfoResult, error) {
	// 1. Try to find track in Postgres database (by UUID)
	var track *model.Track
	if _, err := uuid.Parse(sourcePath); err == nil {
		track, _ = s.Repo.GetTrackByID(ctx, sourcePath)
	}

	// 2. Dual-Engine Option A: Pre-hosted Supabase MP3 CDN URL
	if track != nil && track.AudioFileURL != "" && strings.HasPrefix(track.AudioFileURL, "http") {
		s.recordListeningHistory(ctx, sourcePath, contentType, contentId)
		durationMs := track.Duration * 1000
		return &PlaybackInfoResult{
			HlsURL:      track.AudioFileURL,
			Mp3Fallback: track.AudioFileURL,
			DurationMs:  &durationMs,
		}, nil
	}

	// 3. Dual-Engine Option B: On-Demand YouTube 192 kbps Live Stream
	youtubeID := ""
	if track != nil && track.YoutubeID != nil && *track.YoutubeID != "" {
		youtubeID = *track.YoutubeID
	} else if track == nil {
		cleanedID := extractYoutubeID(sourcePath)
		if len(cleanedID) == 11 {
			youtubeID = cleanedID
		}
	}

	if youtubeID != "" {
		if track != nil {
			s.recordListeningHistory(ctx, sourcePath, contentType, contentId)
		}

		var durationMs *int
		if track != nil && track.Duration > 0 {
			d := track.Duration * 1000
			durationMs = &d
		}

		hlsURL := fmt.Sprintf("/v1/stream/youtube/%s.m3u8", youtubeID)
		mp3URL := fmt.Sprintf("/v1/stream/youtube/%s", youtubeID)
		return &PlaybackInfoResult{
			HlsURL:      hlsURL,
			Mp3Fallback: mp3URL,
			DurationMs:  durationMs,
		}, nil
	}

	// 4. Fall back to local HLS conversion if running locally with disk files
	if trackIdUuid, err := uuid.Parse(sourcePath); err == nil {
		mp3Path, err := s.FileController.GetTrackPath(ctx, trackIdUuid)
		if err == nil && !s.FileController.HLSExists(sourcePath) {
			_, _ = s.FileController.ConvertToHLS(ctx, mp3Path, sourcePath)
		}

		s.recordListeningHistory(ctx, sourcePath, contentType, contentId)

		return &PlaybackInfoResult{
			HlsURL: fmt.Sprintf("/hls/%s/playlist.m3u8", sourcePath),
		}, nil
	}

	return nil, fmt.Errorf("track not found or unplayable: %s", sourcePath)
}

// ResolveYoutubeStreamURL extracts or fetches a cached live 192kbps audio stream URL for a given YouTube ID
func (s *Service) ResolveYoutubeStreamURL(ctx context.Context, rawYoutubeID string) (string, error) {
	youtubeID := extractYoutubeID(rawYoutubeID)
	if youtubeID == "" {
		return "", fmt.Errorf("empty or invalid youtube id")
	}

	cacheKey := fmt.Sprintf("stream:yt:%s", youtubeID)

	// 1. Check L1 in-memory cache (< 0.1ms)
	if val, ok := s.memoryCache.Load(cacheKey); ok {
		if entry, ok := val.(cachedStream); ok {
			if time.Now().Before(entry.expiresAt) {
				return entry.url, nil
			}
			s.memoryCache.Delete(cacheKey)
		}
	}

	// 2. Check L2 Redis cache (< 5ms)
	if s.Redis != nil {
		if cachedURL, err := s.Redis.Get(ctx, cacheKey).Result(); err == nil && cachedURL != "" {
			s.memoryCache.Store(cacheKey, cachedStream{
				url:       cachedURL,
				expiresAt: time.Now().Add(4 * time.Hour),
			})
			return cachedURL, nil
		}
	}

	// 3. Resolve on-demand using yt-dlp
	ytDlp := resolveYtDlpPath()

	execCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	args := []string{
		"-4",
		"-g",
		"--no-playlist",
		"--socket-timeout", "15",
		"-f", "bestaudio[abr<=192]/bestaudio/best",
		"--no-warnings",
	}

	var hasCookies bool
	if cookiesB64 := os.Getenv("YOUTUBE_COOKIES_BASE64"); cookiesB64 != "" {
		cookieFile := filepath.Join(os.TempDir(), "yt_cookies.txt")
		clean := strings.TrimSpace(cookiesB64)
		clean = strings.ReplaceAll(clean, " ", "")
		clean = strings.ReplaceAll(clean, "\n", "")
		clean = strings.ReplaceAll(clean, "\r", "")
		clean = strings.ReplaceAll(clean, "\t", "")
		data, err := base64.StdEncoding.DecodeString(clean)
		if err != nil {
			data, err = base64.RawStdEncoding.DecodeString(clean)
		}
		if err != nil {
			fmt.Printf("[STREAM] error decoding YOUTUBE_COOKIES_BASE64: %v\n", err)
		} else if len(data) > 0 {
			if writeErr := os.WriteFile(cookieFile, data, 0o600); writeErr == nil {
				args = append(args, "--cookies", cookieFile)
				hasCookies = true
				fmt.Printf("[STREAM] successfully loaded %d bytes of YouTube cookies to %s\n", len(data), cookieFile)
			} else {
				fmt.Printf("[STREAM] failed to write cookie file: %v\n", writeErr)
			}
		}
	} else if cookiesPath := os.Getenv("YOUTUBE_COOKIES_PATH"); cookiesPath != "" {
		args = append(args, "--cookies", cookiesPath)
		hasCookies = true
	} else if _, err := os.Stat("cookies.txt"); err == nil {
		args = append(args, "--cookies", "cookies.txt")
		hasCookies = true
	}

	if hasCookies {
		args = append(args, "--extractor-args", "youtube:player_client=mweb,web_embedded")
	} else {
		args = append(args, "--extractor-args", "youtube:player_client=android,mweb,web_embedded")
	}

	args = append(args, youtubeID)

	cmd := exec.CommandContext(execCtx, ytDlp, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("yt-dlp extract stream failed: %w: %s", err, stderr.String())
	}

	rawOutput := strings.TrimSpace(stdout.String())
	if rawOutput == "" {
		return "", fmt.Errorf("yt-dlp returned empty stream url: %s", stderr.String())
	}

	lines := strings.Split(rawOutput, "\n")
	streamURL := strings.TrimSpace(lines[0])

	// 4. Cache for 4 hours (YouTube signed CDN URLs expire after ~6 hours)
	ttl := 4 * time.Hour
	s.memoryCache.Store(cacheKey, cachedStream{
		url:       streamURL,
		expiresAt: time.Now().Add(ttl),
	})

	if s.Redis != nil {
		_ = s.Redis.Set(ctx, cacheKey, streamURL, ttl).Err()
	}

	return streamURL, nil
}

func (s *Service) resolveYoutubeStreamURL(ctx context.Context, rawYoutubeID string) (string, error) {
	return s.ResolveYoutubeStreamURL(ctx, rawYoutubeID)
}

// InvalidateYoutubeStreamURL purges the cached stream URL for a given YouTube ID
func (s *Service) InvalidateYoutubeStreamURL(ctx context.Context, rawYoutubeID string) {
	youtubeID := extractYoutubeID(rawYoutubeID)
	if youtubeID == "" {
		return
	}

	cacheKey := fmt.Sprintf("stream:yt:%s", youtubeID)
	s.memoryCache.Delete(cacheKey)
	if s.Redis != nil {
		_ = s.Redis.Del(ctx, cacheKey).Err()
	}
}

// ExtractYoutubeID sanitizes an input that may be a raw 11-char ID or a YouTube URL
func ExtractYoutubeID(input string) string {
	return extractYoutubeID(input)
}

// extractYoutubeID sanitizes an input that may be a raw 11-char ID or a YouTube URL
func extractYoutubeID(input string) string {
	input = strings.TrimSpace(input)
	input = strings.TrimSuffix(input, ".m3u8")
	if len(input) == 11 && !strings.Contains(input, "/") && !strings.Contains(input, "?") {
		return input
	}
	if strings.Contains(input, "v=") {
		parts := strings.Split(input, "v=")
		if len(parts) > 1 {
			id := parts[1]
			if idx := strings.IndexAny(id, "&?/#"); idx != -1 {
				id = id[:idx]
			}
			return id
		}
	}
	if strings.Contains(input, "youtu.be/") {
		parts := strings.Split(input, "youtu.be/")
		if len(parts) > 1 {
			id := parts[1]
			if idx := strings.IndexAny(id, "&?/#"); idx != -1 {
				id = id[:idx]
			}
			return id
		}
	}
	return input
}

// resolveYtDlpPath checks environment and candidate system paths for yt-dlp
func resolveYtDlpPath() string {
	if envVal := os.Getenv("YT_DLP_PATH"); envVal != "" {
		if path, err := exec.LookPath(envVal); err == nil {
			return path
		}
		if fi, err := os.Stat(envVal); err == nil && !fi.IsDir() {
			return envVal
		}
	}
	candidates := []string{
		"yt-dlp",
		"/opt/homebrew/bin/yt-dlp",
		"/usr/local/bin/yt-dlp",
		"/usr/bin/yt-dlp",
		"./bin/yt-dlp",
		"bin/yt-dlp",
	}
	for _, cand := range candidates {
		if path, err := exec.LookPath(cand); err == nil {
			return path
		}
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			return cand
		}
	}
	return "yt-dlp"
}

func (s *Service) recordListeningHistory(ctx context.Context, sourcePath, contentType, contentId string) {
	if _, err := uuid.Parse(sourcePath); err != nil {
		return
	}

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
