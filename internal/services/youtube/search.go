package youtube

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
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

	redisdb "vvinamp/database/redis"
)

type SearchService struct {
	ytDlpPath string
	apiKey    string
	redisConn *redisdb.Connection
}

func NewSearchService(ytDlpPath string, redisConn *redisdb.Connection) *SearchService {
	return &SearchService{
		ytDlpPath: ytDlpPath,
		apiKey:    os.Getenv("YT_API_KEY"),
		redisConn: redisConn,
	}
}

func (s *SearchService) Search(ctx context.Context, query string) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if s.apiKey == "" {
		return s.SearchYoutubeFromYtDlp(ctx, query)
	}

	results, err := s.SearchYoutubeFromAPI(ctx, query)
	if err == nil {
		return results, nil
	}

	return s.SearchYoutubeFromYtDlp(ctx, query)
}

func (s *SearchService) SearchYoutubeFromAPI(ctx context.Context, query string) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// default to non-paginated (first page)
	results, _, _, err := s.SearchYoutubeFromAPIPaginatedWithPageToken(ctx, query, 0, 5, "", nil)
	return results, err
}

func (s *SearchService) SearchYoutubeFromAPIPaginated(ctx context.Context, query string, offset, limit int) ([]SearchResult, error) {
	results, _, _, err := s.SearchYoutubeFromAPIPaginatedWithPageToken(ctx, query, offset, limit, "", nil)
	return results, err
}

func (s *SearchService) SearchYoutubeFromAPIPaginatedWithPageToken(ctx context.Context, query string, offset, limit int, pageToken string, seenIDs []string) ([]SearchResult, string, []string, error) {
	if strings.TrimSpace(query) == "" {
		return nil, "", nil, fmt.Errorf("search query cannot be empty")
	}

	if s.apiKey == "" {
		return nil, "", nil, fmt.Errorf("YT_API_KEY is not configured; YouTube API search is required")
	}

	if limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}

	seenSet := make(map[string]struct{}, len(seenIDs)+limit)
	for _, id := range seenIDs {
		if id == "" {
			continue
		}
		seenSet[id] = struct{}{}
	}

	results := make([]SearchResult, 0, limit)
	currentPageToken := pageToken
	nextPageToken := ""
	for len(results) < limit {
		pageResults, token, err := s.searchWithYouTubeAPIWithPageToken(ctx, query, currentPageToken, limit)
		if err != nil {
			return nil, "", nil, err
		}

		pageSeen := make(map[string]struct{}, len(pageResults))
		for _, result := range pageResults {
			if result.ID == "" {
				continue
			}
			if _, alreadySeen := seenSet[result.ID]; alreadySeen {
				continue
			}
			if _, alreadyInPage := pageSeen[result.ID]; alreadyInPage {
				continue
			}
			pageSeen[result.ID] = struct{}{}
			results = append(results, result)
			if len(results) >= limit {
				break
			}
		}

		nextPageToken = token
		if token == "" || token == currentPageToken {
			break
		}
		currentPageToken = token
	}

	filteredResults := paginateSearchResults(results, 0, limit, seenSet)
	for _, result := range filteredResults {
		seenSet[result.ID] = struct{}{}
	}

	return filteredResults, nextPageToken, collectSeenIDs(seenSet), nil
}

func collectSeenIDs(seenSet map[string]struct{}) []string {
	ids := make([]string, 0, len(seenSet))
	for id := range seenSet {
		ids = append(ids, id)
	}
	return ids
}

func paginateSearchResults(results []SearchResult, offset, limit int, seenSet map[string]struct{}) []SearchResult {
	if limit <= 0 {
		limit = 5
	}
	if offset < 0 {
		offset = 0
	}

	seen := make(map[string]struct{}, len(results))
	uniqueResults := make([]SearchResult, 0, len(results))
	for _, result := range results {
		if result.ID == "" {
			continue
		}
		if _, exists := seen[result.ID]; exists {
			continue
		}
		if seenSet != nil {
			if _, exists := seenSet[result.ID]; exists {
				continue
			}
		}
		seen[result.ID] = struct{}{}
		uniqueResults = append(uniqueResults, result)
	}

	if offset >= len(uniqueResults) {
		return []SearchResult{}
	}

	end := offset + limit
	if end > len(uniqueResults) {
		end = len(uniqueResults)
	}

	return uniqueResults[offset:end]
}

func (s *SearchService) searchWithYouTubeAPIWithMax(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	endpoint := "https://www.googleapis.com/youtube/v3/search"
	params := neturl.Values{}
	params.Set("part", "snippet")
	params.Set("q", query)
	params.Set("type", "video")
	params.Set("maxResults", fmt.Sprintf("%d", maxResults))
	params.Set("key", s.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("youtube api request creation failed: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusForbidden && isQuotaExceeded(body) {
			log.Println("YouTube Data API daily quota exceeded; auto-switching to optimized yt-dlp fallback")
			fallbackResults, fallbackErr := s.SearchYoutubeFromYtDlpPaginated(ctx, query, 0, maxResults)
			if fallbackErr != nil {
				return nil, fmt.Errorf("yt-dlp fallback after quotaExceeded failed: %w", fallbackErr)
			}
			markFallbackResults(fallbackResults)
			return fallbackResults, nil
		}
		return nil, fmt.Errorf("youtube api returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var searchResp struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title        string `json:"title"`
				ChannelTitle string `json:"channelTitle"`
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

	results, err := s.fetchVideoDetails(ctx, ids)
	if err != nil {
		if isQuotaExceededError(err) {
			log.Println("YouTube Data API daily quota exceeded; auto-switching to optimized yt-dlp fallback")
			fallbackResults, fallbackErr := s.SearchYoutubeFromYtDlpPaginated(ctx, query, 0, maxResults)
			if fallbackErr != nil {
				return nil, fmt.Errorf("yt-dlp fallback after quotaExceeded failed: %w", fallbackErr)
			}
			markFallbackResults(fallbackResults)
			return fallbackResults, nil
		}
		return nil, err
	}

	return results, nil
}

func (s *SearchService) SearchYoutubeFromYtDlp(ctx context.Context, query string) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	return s.SearchYoutubeFromYtDlpPaginated(ctx, query, 0, 5)
}

func (s *SearchService) searchWithYouTubeAPIWithPageToken(ctx context.Context, query, pageToken string, maxResults int) ([]SearchResult, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	endpoint := "https://www.googleapis.com/youtube/v3/search"
	params := neturl.Values{}
	params.Set("part", "snippet")
	params.Set("q", query)
	params.Set("type", "video")
	params.Set("maxResults", fmt.Sprintf("%d", maxResults))
	params.Set("key", s.apiKey)
	if pageToken != "" {
		params.Set("pageToken", pageToken)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("youtube api request creation failed: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("youtube api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusForbidden && isQuotaExceeded(body) {
			log.Println("YouTube Data API daily quota exceeded; auto-switching to optimized yt-dlp fallback")
			fallbackResults, fallbackErr := s.SearchYoutubeFromYtDlpPaginated(ctx, query, 0, maxResults)
			if fallbackErr != nil {
				return nil, "", fmt.Errorf("yt-dlp fallback after quotaExceeded failed: %w", fallbackErr)
			}
			markFallbackResults(fallbackResults)
			return fallbackResults, "", nil
		}
		return nil, "", fmt.Errorf("youtube api returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var searchResp struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title        string `json:"title"`
				ChannelTitle string `json:"channelTitle"`
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
		} `json:"items"`
		NextPageToken string `json:"nextPageToken"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, "", fmt.Errorf("youtube api response decode failed: %w", err)
	}

	ids := make([]string, 0, len(searchResp.Items))
	for _, item := range searchResp.Items {
		if item.ID.VideoID != "" {
			ids = append(ids, item.ID.VideoID)
		}
	}

	if len(ids) == 0 {
		return []SearchResult{}, searchResp.NextPageToken, nil
	}

	results, err := s.fetchVideoDetails(ctx, ids)
	if err != nil {
		if isQuotaExceededError(err) {
			log.Println("YouTube Data API daily quota exceeded; auto-switching to optimized yt-dlp fallback")
			fallbackResults, fallbackErr := s.SearchYoutubeFromYtDlpPaginated(ctx, query, 0, maxResults)
			if fallbackErr != nil {
				return nil, "", fmt.Errorf("yt-dlp fallback after quotaExceeded failed: %w", fallbackErr)
			}
			markFallbackResults(fallbackResults)
			return fallbackResults, "", nil
		}
		return nil, "", err
	}

	return results, searchResp.NextPageToken, nil
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

func (s *SearchService) searchWithYouTubeAPI(ctx context.Context, query string) ([]SearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	endpoint := "https://www.googleapis.com/youtube/v3/search"
	params := neturl.Values{}
	params.Set("part", "snippet")
	params.Set("q", query)
	params.Set("type", "video")
	// params.Set("videoCategoryId", "10")
	params.Set("maxResults", "5")
	params.Set("key", s.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("youtube api request creation failed: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusForbidden && isQuotaExceeded(body) {
			log.Println("YouTube Data API daily quota exceeded; auto-switching to optimized yt-dlp fallback")
			fallbackResults, fallbackErr := s.SearchYoutubeFromYtDlpPaginated(ctx, query, 0, 5)
			if fallbackErr != nil {
				return nil, fmt.Errorf("yt-dlp fallback after quotaExceeded failed: %w", fallbackErr)
			}
			markFallbackResults(fallbackResults)
			return fallbackResults, nil
		}
		return nil, fmt.Errorf("youtube api returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var searchResp struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title        string `json:"title"`
				ChannelTitle string `json:"channelTitle"`
				Thumbnails   struct {
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
					Medium struct {
						URL string `json:"url"`
					} `json:"medium"`
				} `json:"thumbnails"`
			} `json:"snippet"`
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

	results, err := s.fetchVideoDetails(ctx, ids)
	if err != nil {
		if isQuotaExceededError(err) {
			log.Println("YouTube Data API daily quota exceeded; auto-switching to optimized yt-dlp fallback")
			fallbackResults, fallbackErr := s.SearchYoutubeFromYtDlpPaginated(ctx, query, 0, 5)
			if fallbackErr != nil {
				return nil, fmt.Errorf("yt-dlp fallback after quotaExceeded failed: %w", fallbackErr)
			}
			markFallbackResults(fallbackResults)
			return fallbackResults, nil
		}
		return nil, err
	}

	return results, nil
}

func (s *SearchService) fetchVideoDetails(ctx context.Context, ids []string) ([]SearchResult, error) {
	endpoint := "https://www.googleapis.com/youtube/v3/videos"
	params := neturl.Values{}
	params.Set("part", "snippet,contentDetails,statistics")
	params.Set("id", strings.Join(ids, ","))
	params.Set("key", s.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("youtube api details request failed: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube api details request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusForbidden && isQuotaExceeded(body) {
			return nil, fmt.Errorf("quotaExceeded: %s", strings.TrimSpace(string(body)))
		}
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

type googleAPIErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Errors  []struct {
			Reason  string `json:"reason"`
			Message string `json:"message"`
		} `json:"errors"`
	} `json:"error"`
}

func isQuotaExceeded(body []byte) bool {
	var errResp googleAPIErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil {
		for _, e := range errResp.Error.Errors {
			if e.Reason == "quotaExceeded" {
				return true
			}
		}
	}
	return strings.Contains(string(body), "quotaExceeded")
}

func isQuotaExceededError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "quotaExceeded")
}

func markFallbackResults(results []SearchResult) {
	for i := range results {
		results[i].UploaderID = "fallback:yt-dlp"
	}
}

func (s *SearchService) searchWithYtDlp(ctx context.Context, query string, offset, limit int) ([]SearchResult, error) {
	// If cached in Redis, read and slice
	log.Printf("joel123 Performing yt-dlp search for query: %s", query)
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

	// Calculate target count for dynamic limit
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

	// Not cached: run yt-dlp search (--flat-playlist, dynamic limit) and stream results.
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.ytDlpPath,
		"--dump-json",
		"--flat-playlist",
		"--skip-download",
		fmt.Sprintf("ytsearch%d:%s", targetLimit, query),
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
			}
			if err := json.Unmarshal([]byte(line), &raw); err != nil {
				// skip malformed lines
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
