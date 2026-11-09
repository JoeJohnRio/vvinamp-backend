package youtube

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeSearchCache struct {
	results   []SearchResult
	fetchedAt time.Time
	found     bool
	saved     []SearchResult
}

func (c *fakeSearchCache) load(context.Context, string) ([]SearchResult, time.Time, bool) {
	return c.results, c.fetchedAt, c.found
}

func (c *fakeSearchCache) save(_ context.Context, _ string, results []SearchResult) {
	c.saved = results
}

func cachedAgo(age time.Duration) *fakeSearchCache {
	return &fakeSearchCache{results: []SearchResult{{ID: "cached"}}, fetchedAt: time.Now().Add(-age), found: true}
}

type fetchStub struct {
	calls   int
	results []SearchResult
	err     error
}

func (f *fetchStub) fetch(context.Context, string) ([]SearchResult, error) {
	f.calls++
	return f.results, f.err
}

func newTestSearchService(cache searchCache, api, ytDlp *fetchStub) *SearchService {
	return &SearchService{apiKey: "test", cache: cache, fetchAPI: api.fetch, fetchYtDlp: ytDlp.fetch}
}

func TestSearchCachedUsesFreshCacheWithoutFetching(t *testing.T) {
	api, ytDlp := &fetchStub{}, &fetchStub{}
	got, err := newTestSearchService(cachedAgo(time.Hour), api, ytDlp).SearchCached(context.Background(), "song")
	if err != nil || len(got) != 1 || got[0].ID != "cached" {
		t.Fatalf("expected cached result, got %v, %v", got, err)
	}
	if api.calls != 0 || ytDlp.calls != 0 {
		t.Fatalf("expected no fetches, got api=%d ytDlp=%d", api.calls, ytDlp.calls)
	}
}

func TestSearchCachedSavesAPIResults(t *testing.T) {
	cache := &fakeSearchCache{}
	api, ytDlp := &fetchStub{results: []SearchResult{{ID: "api"}}}, &fetchStub{}
	got, err := newTestSearchService(cache, api, ytDlp).SearchCached(context.Background(), "song")
	if err != nil || got[0].ID != "api" {
		t.Fatalf("expected api result, got %v, %v", got, err)
	}
	if ytDlp.calls != 0 || len(cache.saved) != 1 {
		t.Fatalf("expected api result saved without yt-dlp, got ytDlp=%d saved=%v", ytDlp.calls, cache.saved)
	}
}

func TestSearchCachedFallsBackToYtDlp(t *testing.T) {
	api := &fetchStub{err: errors.New("quotaExceeded")}
	ytDlp := &fetchStub{results: []SearchResult{{ID: "ytdlp"}}}
	got, err := newTestSearchService(&fakeSearchCache{}, api, ytDlp).SearchCached(context.Background(), "song")
	if err != nil || got[0].ID != "ytdlp" {
		t.Fatalf("expected yt-dlp result, got %v, %v", got, err)
	}
}

func TestSearchCachedServesStaleCacheWhenSourcesFail(t *testing.T) {
	api, ytDlp := &fetchStub{err: errors.New("down")}, &fetchStub{err: errors.New("blocked")}
	got, err := newTestSearchService(cachedAgo(10*24*time.Hour), api, ytDlp).SearchCached(context.Background(), "song")
	if err != nil || got[0].ID != "cached" {
		t.Fatalf("expected stale cache, got %v, %v", got, err)
	}
}

func TestSearchCachedIgnoresExpiredCache(t *testing.T) {
	api, ytDlp := &fetchStub{err: errors.New("down")}, &fetchStub{err: errors.New("blocked")}
	_, err := newTestSearchService(cachedAgo(31*24*time.Hour), api, ytDlp).SearchCached(context.Background(), "song")
	if !errors.Is(err, errSearchUnavailable) {
		t.Fatalf("expected errSearchUnavailable, got %v", err)
	}
}

func TestSearchCachedHidesInternalErrors(t *testing.T) {
	api, ytDlp := &fetchStub{err: errors.New("key=secret")}, &fetchStub{err: errors.New("stderr")}
	_, err := newTestSearchService(&fakeSearchCache{}, api, ytDlp).SearchCached(context.Background(), "song")
	if !errors.Is(err, errSearchUnavailable) {
		t.Fatalf("expected errSearchUnavailable, got %v", err)
	}
}

func TestSearchCachedRejectsEmptyQuery(t *testing.T) {
	if _, err := newTestSearchService(&fakeSearchCache{}, &fetchStub{}, &fetchStub{}).SearchCached(context.Background(), "   "); err == nil {
		t.Fatal("expected an error for an empty query")
	}
}

func TestNormalizeSearchQuery(t *testing.T) {
	if got := normalizeSearchQuery("  Michael   JACKSON "); got != "michael jackson" {
		t.Fatalf("got %q", got)
	}
}

func TestSearchPage(t *testing.T) {
	results := make([]SearchResult, 12)
	for _, tt := range []struct {
		offset, limit, wantLen int
		wantNext               bool
	}{
		{0, 5, 5, true},
		{5, 5, 5, true},
		{10, 5, 2, false},
		{12, 5, 0, false},
		{-1, 0, 5, true},
	} {
		page, next := SearchPage(results, tt.offset, tt.limit)
		if len(page) != tt.wantLen || next != tt.wantNext {
			t.Errorf("SearchPage(%d, %d) = %d, %v; want %d, %v", tt.offset, tt.limit, len(page), next, tt.wantLen, tt.wantNext)
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
