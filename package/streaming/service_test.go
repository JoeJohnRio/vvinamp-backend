package streaming

import (
	"context"
	"testing"

	"vvinamp/graphql/model"
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
		{"m3u8 suffixed ID", "dQw4w9WgXcQ.m3u8", "dQw4w9WgXcQ"},
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

func TestPlayerQueueWorksWithoutRedis(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()

	empty, err := svc.GetPlayerQueue(ctx, "user")
	if err != nil {
		t.Fatalf("expected an empty queue, got error: %v", err)
	}
	if len(empty.UpcomingTrack) != 0 {
		t.Fatalf("expected no upcoming tracks, got %d", len(empty.UpcomingTrack))
	}

	queue := &model.PlayerQueue{
		UpcomingTrack: []*model.TrackMinimal{{TrackID: "next"}},
		PreviousTrack: []*model.TrackMinimal{},
	}
	if err := svc.SetPlayerQueue(ctx, "user", queue); err != nil {
		t.Fatalf("failed to save queue: %v", err)
	}

	saved, err := svc.GetPlayerQueue(ctx, "user")
	if err != nil {
		t.Fatalf("failed to load queue: %v", err)
	}
	if len(saved.UpcomingTrack) != 1 || saved.UpcomingTrack[0].TrackID != "next" {
		t.Fatalf("expected saved queue, got %+v", saved.UpcomingTrack)
	}
}
