package youtube

import "time"

type SearchResult struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	URL            string `json:"webpage_url"`
	WebpageURL     string `json:"webpage_url"`
	Duration       int    `json:"duration"`
	DurationString string `json:"duration_string"`
	ViewCount      int64  `json:"view_count"`
	Channel        string `json:"uploader"`
	Thumbnail      string `json:"thumbnail"`
}

type DownloadTask struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	YoutubeURL string    `json:"youtube_url"`
	Status     string    `json:"status"`
	Progress   int       `json:"progress"`
	FilePath   string    `json:"file_path,omitempty"`
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DownloadRequest struct {
	YoutubeURL string `json:"youtube_url"`
}
