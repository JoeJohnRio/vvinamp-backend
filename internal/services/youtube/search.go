package youtube

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	neturl "net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"vvinamp/database/postgre"
	redisdb "vvinamp/database/redis"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	searchCacheFreshFor = 7 * 24 * time.Hour
	// YouTube's API policies require stored API data to be refreshed or removed within 30 days.
	searchCacheMaxAge = 30 * 24 * time.Hour
	apiSearchSize     = 50
	ytDlpSearchSize   = 20
	defaultPageSize   = 5
)

var errSearchUnavailable = errors.New("youtube search is unavailable right now")

type searchCache interface {
	load(ctx context.Context, key string) (results []SearchResult, fetchedAt time.Time, found bool)
	save(ctx context.Context, key string, results []SearchResult)
}

type SearchService struct {
	ytDlpPath  string
	apiKey     string
	redisConn  *redisdb.Connection
	cache      searchCache
	fetchAPI   func(ctx context.Context, query string) ([]SearchResult, error)
	fetchYtDlp func(ctx context.Context, query string) ([]SearchResult, error)
}

func NewSearchService(ytDlpPath string, redisConn *redisdb.Connection, postgres *postgre.Connection) *SearchService {
	s := &SearchService{
		ytDlpPath: ytDlpPath,
		apiKey:    os.Getenv("YT_API_KEY"),
		redisConn: redisConn,
		cache:     newPostgresSearchCache(postgres),
	}
	s.fetchAPI = s.searchYouTubeAPI
	s.fetchYtDlp = func(ctx context.Context, query string) ([]SearchResult, error) {
		return s.searchWithYtDlp(ctx, query, 0, ytDlpSearchSize)
	}
	return s
}

// Search returns the first results for query.
func (s *SearchService) Search(ctx context.Context, query string) ([]SearchResult, error) {
	results, err := s.SearchCached(ctx, query)
	if err != nil {
		return nil, err
	}
	page, _ := SearchPage(results, 0, defaultPageSize)
	return page, nil
}

// SearchCached returns the full result list for query, using the first source that has results:
// fresh cache, YouTube API, yt-dlp, then stale cache.
func (s *SearchService) SearchCached(ctx context.Context, query string) ([]SearchResult, error) {
	key := normalizeSearchQuery(query)
	if key == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	cached, fetchedAt, found := s.cache.load(ctx, key)
	age := time.Since(fetchedAt)
	if found && age < searchCacheFreshFor {
		return cached, nil
	}

	results, err := s.fetchFresh(ctx, query)
	if len(results) > 0 {
		s.cache.save(ctx, key, results)
		return results, nil
	}

	if found && age < searchCacheMaxAge {
		return cached, nil
	}
	return results, err
}

func (s *SearchService) fetchFresh(ctx context.Context, query string) ([]SearchResult, error) {
	var apiErr error
	if s.apiKey != "" {
		var results []SearchResult
		results, apiErr = s.fetchAPI(ctx, query)
		if len(results) > 0 {
			return results, nil
		}
		if apiErr != nil {
			log.Printf("youtube api search failed, trying yt-dlp: %v", apiErr)
		}
	}

	results, ytDlpErr := s.fetchYtDlp(ctx, query)
	if len(results) > 0 {
		return results, nil
	}
	if ytDlpErr != nil {
		log.Printf("yt-dlp search failed: %v", ytDlpErr)
	}

	if apiErr != nil || ytDlpErr != nil {
		return nil, errSearchUnavailable
	}
	return []SearchResult{}, nil
}

// SearchPage returns one page of results and whether more results follow it.
func SearchPage(results []SearchResult, offset, limit int) ([]SearchResult, bool) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = defaultPageSize
	}
	if offset >= len(results) {
		return []SearchResult{}, false
	}

	end := offset + limit
	if end > len(results) {
		end = len(results)
	}
	return results[offset:end], end < len(results)
}

func normalizeSearchQuery(query string) string {
	return strings.Join(strings.Fields(strings.ToLower(query)), " ")
}

type youtubeSearchCache struct {
	QueryKey  string    `gorm:"primaryKey;column:query_key"`
	Results   string    `gorm:"type:text;not null"`
	FetchedAt time.Time `gorm:"not null"`
}

func (youtubeSearchCache) TableName() string {
	return "youtube_search_cache"
}

type postgresSearchCache struct {
	db *gorm.DB
}

func newPostgresSearchCache(postgres *postgre.Connection) *postgresSearchCache {
	db, err := postgres.Client()
	if err == nil {
		err = db.AutoMigrate(&youtubeSearchCache{})
	}
	if err != nil {
		log.Printf("youtube search cache disabled: %v", err)
		return &postgresSearchCache{}
	}
	return &postgresSearchCache{db: db}
}

func (c *postgresSearchCache) load(ctx context.Context, key string) ([]SearchResult, time.Time, bool) {
	if c.db == nil {
		return nil, time.Time{}, false
	}

	var row youtubeSearchCache
	if err := c.db.WithContext(ctx).First(&row, "query_key = ?", key).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("youtube search cache read failed: %v", err)
		}
		return nil, time.Time{}, false
	}

	var results []SearchResult
	if err := json.Unmarshal([]byte(row.Results), &results); err != nil {
		log.Printf("youtube search cache decode failed: %v", err)
		return nil, time.Time{}, false
	}
	return results, row.FetchedAt, true
}

func (c *postgresSearchCache) save(ctx context.Context, key string, results []SearchResult) {
	if c.db == nil {
		return
	}

	encoded, err := json.Marshal(results)
	if err != nil {
		log.Printf("youtube search cache encode failed: %v", err)
		return
	}

	row := youtubeSearchCache{QueryKey: key, Results: string(encoded), FetchedAt: time.Now().UTC()}
	if err := c.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error; err != nil {
		log.Printf("youtube search cache write failed: %v", err)
	}
}

// searchYouTubeAPI spends one search.list call and fetches enough results to serve every page from the cache.
func (s *SearchService) searchYouTubeAPI(ctx context.Context, query string) ([]SearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	params := neturl.Values{}
	params.Set("part", "id")
	params.Set("q", query)
	params.Set("type", "video")
	params.Set("videoCategoryId", "10")
	params.Set("videoEmbeddable", "true")
	params.Set("maxResults", strconv.Itoa(apiSearchSize))
	params.Set("key", s.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/youtube/v3/search?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("youtube api request creation failed: %w", redactURLError(err))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube api request failed: %w", redactURLError(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("youtube api returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var searchResp struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("youtube api response decode failed: %w", err)
	}

	ids := make([]string, 0, len(searchResp.Items))
	for _, item := range searchResp.Items {
		if item.ID.VideoID != "" {
			ids = append(ids, item.ID.VideoID)
		}
	}
	if len(ids) == 0 {
		return []SearchResult{}, nil
	}

	details, err := s.fetchVideoDetails(ctx, ids)
	if err != nil {
		return nil, err
	}

	// videos.list doesn't promise to keep the relevance order from search.list.
	byID := make(map[string]SearchResult, len(details))
	for _, result := range details {
		byID[result.ID] = result
	}
	results := make([]SearchResult, 0, len(details))
	for _, id := range ids {
		if result, ok := byID[id]; ok {
			results = append(results, result)
		}
	}
	return results, nil
}

func (s *SearchService) SearchYoutubeFromYtDlp(ctx context.Context, query string) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	return s.SearchYoutubeFromYtDlpPaginated(ctx, query, 0, 5)
}

func (s *SearchService) SearchYoutubeFromYtDlpPaginated(ctx context.Context, query string, offset, limit int) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if limit <= 0 {
		limit = 5
	}

	return s.searchWithYtDlp(ctx, query, offset, limit)
}

func (s *SearchService) fetchVideoDetails(ctx context.Context, ids []string) ([]SearchResult, error) {
	endpoint := "https://www.googleapis.com/youtube/v3/videos"
	params := neturl.Values{}
	params.Set("part", "snippet,contentDetails,statistics")
	params.Set("id", strings.Join(ids, ","))
	params.Set("key", s.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("youtube api details request failed: %w", redactURLError(err))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube api details request failed: %w", redactURLError(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("youtube api details returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var videoResp struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title        string `json:"title"`
				ChannelTitle string `json:"channelTitle"`
				ChannelID    string `json:"channelId"`
				Thumbnails   struct {
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
					Medium struct {
						URL string `json:"url"`
					} `json:"medium"`
					High struct {
						URL string `json:"url"`
					} `json:"high"`
					Standard struct {
						URL string `json:"url"`
					} `json:"standard"`
					Maxres struct {
						URL string `json:"url"`
					} `json:"maxres"`
				} `json:"thumbnails"`
			} `json:"snippet"`
			ContentDetails struct {
				Duration string `json:"duration"`
			} `json:"contentDetails"`
			Statistics struct {
				ViewCount string `json:"viewCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&videoResp); err != nil {
		return nil, fmt.Errorf("youtube api details decode failed: %w", err)
	}

	results := make([]SearchResult, 0, len(videoResp.Items))
	for _, item := range videoResp.Items {
		duration := parseISO8601Duration(item.ContentDetails.Duration)
		viewCount := int64(0)
		if item.Statistics.ViewCount != "" {
			if parsed, err := strconv.ParseInt(item.Statistics.ViewCount, 10, 64); err == nil {
				viewCount = parsed
			}
		}

		// Prefer highest-resolution thumbnail available
		thumbnail := item.Snippet.Thumbnails.Maxres.URL
		if thumbnail == "" {
			thumbnail = item.Snippet.Thumbnails.High.URL
		}
		if thumbnail == "" {
			thumbnail = item.Snippet.Thumbnails.Medium.URL
		}
		if thumbnail == "" {
			thumbnail = item.Snippet.Thumbnails.Default.URL
		}

		results = append(results, SearchResult{
			ID:             item.ID,
			Title:          item.Snippet.Title,
			URL:            fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ID),
			WebpageURL:     fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ID),
			Duration:       duration,
			DurationString: item.ContentDetails.Duration,
			ViewCount:      viewCount,
			Channel:        item.Snippet.ChannelTitle,
			ChannelID:      item.Snippet.ChannelID,
			UploaderID:     item.Snippet.ChannelID,
			Thumbnail:      thumbnail,
		})
	}

	return results, nil
}

// redactURLError drops the request URL, which contains the API key, from HTTP client errors.
func redactURLError(err error) error {
	var urlErr *neturl.Error
	if errors.As(err, &urlErr) {
		return urlErr.Err
	}
	return err
}

func parseISO8601Duration(value string) int {
	re := regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 4 {
		return 0
	}

	hours, _ := strconv.Atoi(defaultString(matches[1], "0"))
	minutes, _ := strconv.Atoi(defaultString(matches[2], "0"))
	seconds, _ := strconv.Atoi(defaultString(matches[3], "0"))
	return hours*3600 + minutes*60 + seconds
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func isPlayableYoutubeVideo(ieKey, id string) bool {
	if ieKey != "" && ieKey != "Youtube" {
		return false
	}
	if len(id) != 11 || strings.HasPrefix(id, "UC") {
		return false
	}
	return true
}

func (s *SearchService) searchWithYtDlp(ctx context.Context, query string, offset, limit int) ([]SearchResult, error) {
	// If cached in Redis, read and slice
	key := fmt.Sprintf("ytsearch:%x", sha1.Sum([]byte(strings.ToLower(query))))
	if s.redisConn != nil {
		if client, err := s.redisConn.Client(); err == nil {
			if raw, err := client.Get(ctx, key).Result(); err == nil {
				var cached []SearchResult
				if err := json.Unmarshal([]byte(raw), &cached); err == nil {
					if offset >= len(cached) {
						return []SearchResult{}, nil
					}
					end := offset + limit
					if end > len(cached) {
						end = len(cached)
					}
					return cached[offset:end], nil
				}
			}
		}
	}

	// Calculate target count for dynamic limit with extra buffer for filtered non-video items
	targetLimit := offset + limit
	if targetLimit <= 0 {
		targetLimit = limit
	}
	if targetLimit <= 0 {
		targetLimit = 10
	}
	if targetLimit > 50 {
		targetLimit = 50
	}
	fetchCount := targetLimit + 5
	if fetchCount > 50 {
		fetchCount = 50
	}

	// Not cached: run yt-dlp search (--flat-playlist, dynamic limit) and stream results.
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.ytDlpPath,
		"-4",
		"--extractor-args", "youtube:player_client=tv_downgraded,android_vr,android",
		"--dump-json",
		"--flat-playlist",
		"--skip-download",
		fmt.Sprintf("ytsearch%d:%s", fetchCount, query),
	)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("yt-dlp start failed: %w: %s", err, stderr.String())
	}

	deliverCh := make(chan SearchResult, 128)
	errCh := make(chan error, 1)

	results := make([]SearchResult, 0, targetLimit)
	var mu sync.Mutex

	// Single goroutine: scan, append to results, send to deliverCh for immediate consumption,
	// and when finished persist the full results to Redis.
	go func() {
		defer close(deliverCh)
		scanner := bufio.NewScanner(stdoutPipe)
		// increase scanner buffer to handle long JSON lines
		scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}
			var raw struct {
				SearchResult
				ChannelName string `json:"channel"`
				Thumbnails  []struct {
					URL string `json:"url"`
				} `json:"thumbnails"`
				IEKey string `json:"ie_key"`
			}
			if err := json.Unmarshal([]byte(line), &raw); err != nil {
				// skip malformed lines
				continue
			}

			// Filter out non-video entries (e.g. channels, tabs, playlists, or channel IDs starting with UC)
			if !isPlayableYoutubeVideo(raw.IEKey, raw.SearchResult.ID) {
				continue
			}

			r := raw.SearchResult
			if r.Channel == "" && raw.ChannelName != "" {
				r.Channel = raw.ChannelName
			}
			if r.Thumbnail == "" && len(raw.Thumbnails) > 0 {
				r.Thumbnail = raw.Thumbnails[0].URL
			}
			if r.WebpageURL == "" && r.ID != "" {
				r.WebpageURL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", r.ID)
			}
			if r.URL == "" {
				r.URL = r.WebpageURL
			}
			if r.Thumbnail == "" && r.ID != "" {
				r.Thumbnail = fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", r.ID)
			}

			mu.Lock()
			results = append(results, r)
			mu.Unlock()

			// best-effort deliver to caller; non-blocking to avoid stalling scanner
			select {
			case deliverCh <- r:
			default:
			}
		}

		if scErr := scanner.Err(); scErr != nil {
			errCh <- scErr
			return
		}
		if waitErr := cmd.Wait(); waitErr != nil {
			errCh <- fmt.Errorf("yt-dlp wait error: %w: %s", waitErr, stderr.String())
			return
		}

		// Persist complete results to Redis (TTL with jitter)
		if s.redisConn != nil {
			if client, err := s.redisConn.Client(); err == nil {
				mu.Lock()
				b, jerr := json.Marshal(results)
				mu.Unlock()
				if jerr == nil {
					ttl := 10*time.Minute + time.Duration(rand.Intn(60))*time.Second
					_ = client.Set(context.Background(), key, b, ttl).Err()
				}
			}
		}

		errCh <- nil
	}()

	// Collect up to `limit` items to return immediately
	collected := make([]SearchResult, 0, limit)
	for r := range deliverCh {
		collected = append(collected, r)
		if len(collected) >= limit {
			break
		}
	}

	// If we didn't receive enough via deliverCh but scanner appended more, copy them
	if len(collected) < limit {
		mu.Lock()
		available := len(results)
		start := len(collected)
		end := start + (limit - start)
		if end > available {
			end = available
		}
		if start < end {
			collected = append(collected, results[start:end]...)
		}
		mu.Unlock()
	}

	// Check for scanner/cmd errors (non-blocking)
	select {
	case e := <-errCh:
		if e != nil {
			return nil, e
		}
	default:
	}

	if offset >= len(results) {
		return []SearchResult{}, nil
	}
	end := offset + limit
	if end > len(results) {
		end = len(results)
	}
	return results[offset:end], nil
}
