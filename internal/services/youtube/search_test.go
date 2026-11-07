package youtube

import (
	"fmt"
	"testing"
)

func TestPaginateSearchResultsSkipsDuplicates(t *testing.T) {
	results := []SearchResult{
		{ID: "1", Title: "first"},
		{ID: "2", Title: "second"},
		{ID: "2", Title: "second duplicate"},
		{ID: "3", Title: "third"},
		{ID: "4", Title: "fourth"},
	}

	got := paginateSearchResults(results, 0, 2, map[string]struct{}{"2": {}})
	if len(got) != 2 {
		t.Fatalf("expected 2 unique results, got %d", len(got))
	}
	if got[0].ID != "1" || got[1].ID != "3" {
		t.Fatalf("expected page to contain IDs [1 3], got [%s %s]", got[0].ID, got[1].ID)
	}
}

func TestIsQuotaExceeded(t *testing.T) {
	quotaJSON := []byte(`{
		"error": {
			"code": 403,
			"message": "The request cannot be completed because you have exceeded your quota.",
			"errors": [
				{
					"message": "The request cannot be completed because you have exceeded your quota.",
					"domain": "youtube.quota",
					"reason": "quotaExceeded"
				}
			]
		}
	}`)

	if !isQuotaExceeded(quotaJSON) {
		t.Fatalf("expected isQuotaExceeded to return true for standard quotaExceeded JSON")
	}

	rawText := []byte(`403 Forbidden: quotaExceeded limit reached`)
	if !isQuotaExceeded(rawText) {
		t.Fatalf("expected isQuotaExceeded to return true for plain text containing quotaExceeded")
	}

	otherForbidden := []byte(`{
		"error": {
			"code": 403,
			"message": "Access denied.",
			"errors": [
				{
					"reason": "accessDenied"
				}
			]
		}
	}`)
	if isQuotaExceeded(otherForbidden) {
		t.Fatalf("expected isQuotaExceeded to return false for accessDenied")
	}
}

func TestIsQuotaExceededError(t *testing.T) {
	if !isQuotaExceededError(fmt.Errorf("youtube api returned 403: quotaExceeded")) {
		t.Fatalf("expected isQuotaExceededError to return true")
	}
	if isQuotaExceededError(fmt.Errorf("generic network timeout")) {
		t.Fatalf("expected isQuotaExceededError to return false")
	}
	if isQuotaExceededError(nil) {
		t.Fatalf("expected isQuotaExceededError to return false for nil error")
	}
}

func TestMarkFallbackResults(t *testing.T) {
	results := []SearchResult{
		{ID: "vid1", Title: "Title 1"},
		{ID: "vid2", Title: "Title 2", UploaderID: "original"},
	}
	markFallbackResults(results)
	for i, r := range results {
		if r.UploaderID != "fallback:yt-dlp" {
			t.Fatalf("result[%d] UploaderID expected 'fallback:yt-dlp', got '%s'", i, r.UploaderID)
		}
	}
}

func TestIsPlayableYoutubeVideo(t *testing.T) {
	tests := []struct {
		name     string
		ieKey    string
		id       string
		expected bool
	}{
		{"valid video", "Youtube", "dQw4w9WgXcQ", true},
		{"valid video empty ieKey", "", "dQw4w9WgXcQ", true},
		{"channel tab", "YoutubeTab", "UCcBYgbqKS2l0xkbI2nawDOA", false},
		{"channel with Youtube ieKey", "Youtube", "UCcBYgbqKS2l0xkbI2nawDOA", false},
		{"playlist", "YoutubePlaylist", "PL1234567890", false},
		{"short id", "Youtube", "short", false},
		{"long id", "Youtube", "123456789012", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPlayableYoutubeVideo(tt.ieKey, tt.id); got != tt.expected {
				t.Errorf("isPlayableYoutubeVideo(%q, %q) = %v, expected %v", tt.ieKey, tt.id, got, tt.expected)
			}
		})
	}
}
