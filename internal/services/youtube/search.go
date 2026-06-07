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
	return s.SearchYoutubeFromAPIPaginated(ctx, query, 0, 5)
}

func (s *SearchService) SearchYoutubeFromAPIPaginated(ctx context.Context, query string, offset, limit int) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if s.apiKey == "" {
		return nil, fmt.Errorf("YT_API_KEY is not configured; YouTube API search is required")
	}

	if limit <= 0 {
		limit = 5
	}

	// YouTube API supports maxResults up to 50
	fetchCount := offset + limit
	if fetchCount < 1 {
		fetchCount = 5
	}
	if fetchCount > 50 {
		fetchCount = 50
	}

	// reuse searchWithYouTubeAPI but request a specific maxResults
	results, err := s.searchWithYouTubeAPIWithMax(ctx, query, fetchCount)
	if err != nil {
		return nil, err
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

	return s.fetchVideoDetails(ctx, ids)
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

	return s.fetchVideoDetails(ctx, ids)
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

	// Not cached: run yt-dlp search (ytsearch50) and stream results.
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.ytDlpPath,
		"--dump-json",
		"--skip-download",
		fmt.Sprintf("ytsearch50:%s", query),
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

	results := make([]SearchResult, 0, 50)
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
			var r SearchResult
			if err := json.Unmarshal([]byte(line), &r); err != nil {
				// skip malformed lines
				continue
			}
			if r.WebpageURL == "" && r.ID != "" {
				r.WebpageURL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", r.ID)
			}
			if r.URL == "" {
				r.URL = r.WebpageURL
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
