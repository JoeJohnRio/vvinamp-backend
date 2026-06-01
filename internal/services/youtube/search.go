package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SearchService struct {
	ytDlpPath string
	apiKey    string
}

func NewSearchService(ytDlpPath string) *SearchService {
	return &SearchService{
		ytDlpPath: ytDlpPath,
		apiKey:    os.Getenv("YT_API_KEY"),
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
		log.Printf("joel123 YouTube API search successful for query: %s", query)
		return results, nil
	}

	log.Printf("joel123 YouTube API search failed for err: %s", err)
	return s.SearchYoutubeFromYtDlp(ctx, query)
}

func (s *SearchService) SearchYoutubeFromAPI(ctx context.Context, query string) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// default to non-paginated (first page)
	return s.SearchYoutubeFromAPIPaginated(ctx, query, 0, 10)
}

func (s *SearchService) SearchYoutubeFromAPIPaginated(ctx context.Context, query string, offset, limit int) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if s.apiKey == "" {
		return nil, fmt.Errorf("YT_API_KEY is not configured; YouTube API search is required")
	}

	if limit <= 0 {
		limit = 10
	}

	// YouTube API supports maxResults up to 50
	fetchCount := offset + limit
	if fetchCount < 1 {
		fetchCount = 10
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
	params := url.Values{}
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

	return s.SearchYoutubeFromYtDlpPaginated(ctx, query, 0, 10)
}

func (s *SearchService) SearchYoutubeFromYtDlpPaginated(ctx context.Context, query string, offset, limit int) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if limit <= 0 {
		limit = 10
	}

	return s.searchWithYtDlp(ctx, query, offset, limit)
}

func (s *SearchService) searchWithYouTubeAPI(ctx context.Context, query string) ([]SearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	endpoint := "https://www.googleapis.com/youtube/v3/search"
	params := url.Values{}
	params.Set("part", "snippet")
	params.Set("q", query)
	params.Set("type", "video")
	// params.Set("videoCategoryId", "10")
	params.Set("maxResults", "10")
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
	params := url.Values{}
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
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	fetchCount := offset + limit
	if fetchCount < 1 {
		fetchCount = 10
	}
	if fetchCount > 100 {
		fetchCount = 100
	}

	searchURL := fmt.Sprintf("ytsearch%d:%s music", fetchCount, query)
	cmd := exec.CommandContext(ctx, s.ytDlpPath,
		"--dump-json",
		"--skip-download",
		searchURL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp search failed: %w: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return []SearchResult{}, nil
	}

	lines := strings.Split(output, "\n")
	results := make([]SearchResult, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var result SearchResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			return nil, fmt.Errorf("failed to parse yt-dlp search result: %w", err)
		}

		// If yt-dlp JSON doesn't include webpage/url fields, construct them from the id.
		if result.WebpageURL == "" && result.ID != "" {
			result.WebpageURL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", result.ID)
		}
		if result.URL == "" {
			result.URL = result.WebpageURL
		}

		results = append(results, result)
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
