package streaming

import (
	"context"
	"testing"
	"time"
)

func TestExtractYoutubeID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain 11-char ID", "dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"standard watch URL", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"watch URL with additional query params", "https://www.youtube.com/watch?v=dQw4w9WgXcQ&feature=share", "dQw4w9WgXcQ"},
		{"youtu.be short URL", "https://youtu.be/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"youtu.be with query", "https://youtu.be/dQw4w9WgXcQ?t=42", "dQw4w9WgXcQ"},
		{"whitespace padded", "  dQw4w9WgXcQ  ", "dQw4w9WgXcQ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractYoutubeID(tt.input)
			if got != tt.expected {
				t.Errorf("extractYoutubeID(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestResolveYtDlpPath(t *testing.T) {
	path := resolveYtDlpPath()
	if path == "" {
		t.Fatalf("expected non-empty yt-dlp path")
	}
}

func TestMemoryCacheHits(t *testing.T) {
	svc := &Service{}
	cacheKey := "stream:yt:dQw4w9WgXcQ"
	expectedURL := "https://googlevideo.com/videoplayback/test123"

	// 1. Prime cache with 1 hour TTL
	svc.memoryCache.Store(cacheKey, cachedStream{
		url:       expectedURL,
		expiresAt: time.Now().Add(1 * time.Hour),
	})

	// 2. Fetch using resolveYoutubeStreamURL
	resolved, err := svc.resolveYoutubeStreamURL(context.Background(), "dQw4w9WgXcQ")
	if err != nil {
		t.Fatalf("unexpected error on cache hit: %v", err)
	}
	if resolved != expectedURL {
		t.Errorf("expected cached url %q, got %q", expectedURL, resolved)
	}
}
