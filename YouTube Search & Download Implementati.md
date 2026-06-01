# YouTube Search & Download Implementation Guide (Go + GraphQL)

## Overview
This guide implements YouTube search and download functionality for your Go/GraphQL project. We'll use `yt-dlp` for extraction and store files in a configurable directory.

---

## 1. Project Structure

```
your-project/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── graph/
│   │   ├── schema.graphqls
│   │   ├── resolver.go
│   │   └── youtube.go          # NEW: YouTube resolvers
│   ├── services/
│   │   └── youtube/
│   │       ├── search.go        # NEW: Search logic
│   │       ├── download.go      # NEW: Download logic
│   │       └── types.go         # NEW: Data types
│   └── config/
│       └── config.go            # NEW: Configuration
├── go.mod
├── go.sum
└── .env                          # NEW: Environment variables
```

---

## 2. Environment Configuration (.env)

```bash
# Storage configuration
STORAGE_PATH=./media/downloads
TEMP_PATH=./media/temp

# YouTube extraction tool path
YT_DLP_PATH=yt-dlp

# API configuration
API_PORT=3001
API_HOST=localhost
CORS_ORIGINS=*

# Download settings
MAX_DOWNLOAD_SIZE=500  # MB
DOWNLOAD_TIMEOUT=600   # seconds
```

---

## 3. Configuration Package

**File: `internal/config/config.go`**

```go
package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	StoragePath    string
	TempPath       string
	YtDlpPath      string
	APIPort        string
	APIHost        string
	MaxDownloadSize int64 // in MB
	DownloadTimeout int   // in seconds
}

func NewConfig() *Config {
	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "./media/downloads"
	}

	tempPath := os.Getenv("TEMP_PATH")
	if tempPath == "" {
		tempPath = "./media/temp"
	}

	ytDlpPath := os.Getenv("YT_DLP_PATH")
	if ytDlpPath == "" {
		ytDlpPath = "yt-dlp"
	}

	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "3001"
	}

	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = "localhost"
	}

	return &Config{
		StoragePath:     ensureDir(storagePath),
		TempPath:        ensureDir(tempPath),
		YtDlpPath:       ytDlpPath,
		APIPort:         apiPort,
		APIHost:         apiHost,
		MaxDownloadSize: 500,
		DownloadTimeout: 600,
	}
}

func ensureDir(path string) string {
	absPath, _ := filepath.Abs(path)
	os.MkdirAll(absPath, 0755)
	return absPath
}
```

---

## 4. YouTube Types

**File: `internal/services/youtube/types.go`**

```go
package youtube

import "time"

// SearchResult represents a YouTube search result
type SearchResult struct {
	Title              string `json:"title"`
	URL                string `json:"webpage_url"`
	WebpageURL         string `json:"webpage_url"`
	Duration           int    `json:"duration"` // in seconds
	DurationString     string `json:"duration_string"`
	ViewCount          int64  `json:"view_count"`
	Channel            string `json:"channel"`
	ChannelIsVerified  *bool  `json:"channel_is_verified"`
	Thumbnail          string `json:"thumbnail"`
	ID                 string `json:"id"`
}

// DownloadTask represents a download task
type DownloadTask struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	YoutubeURL string   `json:"youtube_url"`
	Status    string    `json:"status"` // pending, downloading, completed, failed
	Progress  int       `json:"progress"` // 0-100
	FilePath  string    `json:"file_path"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DownloadRequest is the input for download mutations
type DownloadRequest struct {
	Title      string `json:"title"`
	YoutubeURL string `json:"youtube_url"`
}
```

---

## 5. YouTube Search Service

**File: `internal/services/youtube/search.go`**

```go
package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type SearchService struct {
	ytDlpPath string
}

func NewSearchService(ytDlpPath string) *SearchService {
	return &SearchService{
		ytDlpPath: ytDlpPath,
	}
}

// Search performs a YouTube search
func (s *SearchService) Search(ctx context.Context, query string) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// Limit query length
	if len(query) > 500 {
		query = query[:500]
	}

	// Build yt-dlp command for searching
	// Note: yt-dlp doesn't natively search, so we search via ytsearch: protocol
	searchURL := fmt.Sprintf("ytsearch5:%s", query)

	cmd := exec.CommandContext(ctx, s.ytDlpPath,
		"--dump-json",
		"--no-warnings",
		"--skip-download",
		"-j",
		searchURL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp search failed: %w", err)
	}

	// Parse JSON output
	var results []SearchResult
	output := stdout.String()

	// yt-dlp returns multiple JSON objects separated by newlines
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		var result SearchResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			continue // Skip malformed entries
		}

		// Ensure required fields
		if result.Title == "" || result.WebpageURL == "" {
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

// GetVideoInfo retrieves detailed video information
func (s *SearchService) GetVideoInfo(ctx context.Context, videoURL string) (*SearchResult, error) {
	cmd := exec.CommandContext(ctx, s.ytDlpPath,
		"--dump-json",
		"--no-warnings",
		"--skip-download",
		videoURL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to fetch video info: %w", err)
	}

	var result SearchResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse video info: %w", err)
	}

	return &result, nil
}
```

---

## 6. YouTube Download Service

**File: `internal/services/youtube/download.go`**

```go
package youtube

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DownloadService struct {
	storagePath string
	tempPath    string
	ytDlpPath   string
	timeout     time.Duration
	maxSize     int64

	// In-memory task tracking (use database in production)
	tasks map[string]*DownloadTask
	mu    sync.RWMutex
}

func NewDownloadService(storagePath, tempPath, ytDlpPath string, timeoutSec int, maxSizeMB int64) *DownloadService {
	return &DownloadService{
		storagePath: storagePath,
		tempPath:    tempPath,
		ytDlpPath:   ytDlpPath,
		timeout:     time.Duration(timeoutSec) * time.Second,
		maxSize:     maxSizeMB * 1024 * 1024,
		tasks:       make(map[string]*DownloadTask),
	}
}

// StartDownload initiates an async download task
func (d *DownloadService) StartDownload(ctx context.Context, req DownloadRequest) (*DownloadTask, error) {
	task := &DownloadTask{
		ID:        generateTaskID(),
		Title:     req.Title,
		YoutubeURL: req.YoutubeURL,
		Status:    "pending",
		Progress:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	d.mu.Lock()
	d.tasks[task.ID] = task
	d.mu.Unlock()

	// Start download in background
	go d.executeDownload(context.Background(), task)

	return task, nil
}

// GetTask retrieves task status
func (d *DownloadService) GetTask(taskID string) (*DownloadTask, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	task, exists := d.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found")
	}

	return task, nil
}

// ListTasks returns all tasks
func (d *DownloadService) ListTasks() []*DownloadTask {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var tasks []*DownloadTask
	for _, task := range d.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// executeDownload runs the actual download
func (d *DownloadService) executeDownload(ctx context.Context, task *DownloadTask) {
	defer func() {
		d.mu.Lock()
		task.UpdatedAt = time.Now()
		d.mu.Unlock()
	}()

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	d.updateTask(task.ID, "downloading", 25, "")

	// Sanitize title for filename
	filename := sanitizeFilename(task.Title)
	outputPath := filepath.Join(d.storagePath, filename+".mp3")

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		d.updateTask(task.ID, "completed", 100, "")
		d.mu.Lock()
		task.FilePath = outputPath
		d.mu.Unlock()
		return
	}

	// Build yt-dlp command to download audio
	cmd := exec.CommandContext(ctx, d.ytDlpPath,
		"-f", "bestaudio/best",
		"-x", // Extract audio
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"--no-warnings",
		"--quiet",
		"-o", outputPath,
		task.YoutubeURL,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	d.updateTask(task.ID, "downloading", 50, "")

	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if strings.Contains(errMsg, "context deadline exceeded") {
			errMsg = "Download timeout"
		}
		d.updateTask(task.ID, "failed", 0, errMsg)
		return
	}

	// Verify file was created and check size
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		d.updateTask(task.ID, "failed", 0, "File creation verification failed")
		return
	}

	if fileInfo.Size() > d.maxSize {
		os.Remove(outputPath)
		d.updateTask(task.ID, "failed", 0, "File size exceeds maximum limit")
		return
	}

	d.updateTask(task.ID, "completed", 100, "")
	d.mu.Lock()
	task.FilePath = outputPath
	d.mu.Unlock()
}

// updateTask updates task status (thread-safe)
func (d *DownloadService) updateTask(taskID, status string, progress int, errMsg string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if task, exists := d.tasks[taskID]; exists {
		task.Status = status
		task.Progress = progress
		task.Error = errMsg
		task.UpdatedAt = time.Now()
	}
}

// Helper functions
func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

func sanitizeFilename(filename string) string {
	// Remove invalid filename characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Limit length
	if len(result) > 200 {
		result = result[:200]
	}
	return result
}
```

---

## 7. GraphQL Schema

**File: `internal/graph/schema.graphqls`**

```graphql
type SearchResult {
  id: String!
  title: String!
  url: String!
  duration: Int!
  durationString: String!
  viewCount: Int!
  channel: String!
  channelIsVerified: Boolean
  thumbnail: String!
}

type DownloadTask {
  id: String!
  title: String!
  youtubeUrl: String!
  status: String!
  progress: Int!
  filePath: String
  error: String
  createdAt: String!
  updatedAt: String!
}

input DownloadInput {
  title: String!
  youtubeUrl: String!
}

type Query {
  searchYoutube(query: String!): [SearchResult!]!
  getDownloadTask(id: String!): DownloadTask
  listDownloadTasks: [DownloadTask!]!
}

type Mutation {
  startDownload(input: DownloadInput!): DownloadTask!
}
```

---

## 8. GraphQL Resolvers

**File: `internal/graph/youtube.go`**

```go
package graph

import (
	"context"
	"github.com/yourusername/yourproject/internal/services/youtube"
)

// YoutubeResolvers handles YouTube-related queries and mutations
type YoutubeResolvers struct {
	searchSvc   *youtube.SearchService
	downloadSvc *youtube.DownloadService
}

func NewYoutubeResolvers(searchSvc *youtube.SearchService, downloadSvc *youtube.DownloadService) *YoutubeResolvers {
	return &YoutubeResolvers{
		searchSvc:   searchSvc,
		downloadSvc: downloadSvc,
	}
}

// SearchYoutube resolves the searchYoutube query
func (r *YoutubeResolvers) SearchYoutube(ctx context.Context, query string) ([]*SearchResultResolver, error) {
	results, err := r.searchSvc.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	var resolvers []*SearchResultResolver
	for _, result := range results {
		resolvers = append(resolvers, &SearchResultResolver{result})
	}
	return resolvers, nil
}

// StartDownload resolves the startDownload mutation
func (r *YoutubeResolvers) StartDownload(ctx context.Context, input struct {
	Title      string `json:"title"`
	YoutubeUrl string `json:"youtubeUrl"`
}) (*DownloadTaskResolver, error) {
	req := youtube.DownloadRequest{
		Title:      input.Title,
		YoutubeURL: input.YoutubeUrl,
	}

	task, err := r.downloadSvc.StartDownload(ctx, req)
	if err != nil {
		return nil, err
	}

	return &DownloadTaskResolver{task}, nil
}

// GetDownloadTask resolves getting a task by ID
func (r *YoutubeResolvers) GetDownloadTask(ctx context.Context, id string) (*DownloadTaskResolver, error) {
	task, err := r.downloadSvc.GetTask(id)
	if err != nil {
		return nil, err
	}
	return &DownloadTaskResolver{task}, nil
}

// ListDownloadTasks returns all tasks
func (r *YoutubeResolvers) ListDownloadTasks(ctx context.Context) ([]*DownloadTaskResolver, error) {
	tasks := r.downloadSvc.ListTasks()
	var resolvers []*DownloadTaskResolver
	for _, task := range tasks {
		resolvers = append(resolvers, &DownloadTaskResolver{task})
	}
	return resolvers, nil
}

// SearchResultResolver wraps youtube.SearchResult
type SearchResultResolver struct {
	*youtube.SearchResult
}

func (r *SearchResultResolver) Id() string                 { return r.SearchResult.ID }
func (r *SearchResultResolver) Title() string             { return r.SearchResult.Title }
func (r *SearchResultResolver) Url() string               { return r.SearchResult.WebpageURL }
func (r *SearchResultResolver) Duration() int             { return r.SearchResult.Duration }
func (r *SearchResultResolver) DurationString() string    { return r.SearchResult.DurationString }
func (r *SearchResultResolver) ViewCount() int            { return int(r.SearchResult.ViewCount) }
func (r *SearchResultResolver) Channel() string           { return r.SearchResult.Channel }
func (r *SearchResultResolver) ChannelIsVerified() *bool  { return r.SearchResult.ChannelIsVerified }
func (r *SearchResultResolver) Thumbnail() string         { return r.SearchResult.Thumbnail }

// DownloadTaskResolver wraps youtube.DownloadTask
type DownloadTaskResolver struct {
	*youtube.DownloadTask
}

func (r *DownloadTaskResolver) Id() string          { return r.DownloadTask.ID }
func (r *DownloadTaskResolver) YoutubeUrl() string  { return r.DownloadTask.YoutubeURL }
func (r *DownloadTaskResolver) CreatedAt() string   { return r.DownloadTask.CreatedAt.String() }
func (r *DownloadTaskResolver) UpdatedAt() string   { return r.DownloadTask.UpdatedAt.String() }
```

---

## 9. REST API Endpoints (Fallback)

If you need REST endpoints instead of/alongside GraphQL:

**File: `internal/handlers/youtube.go`**

```go
package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/yourusername/yourproject/internal/services/youtube"
)

type YoutubeHandler struct {
	searchSvc   *youtube.SearchService
	downloadSvc *youtube.DownloadService
}

func NewYoutubeHandler(searchSvc *youtube.SearchService, downloadSvc *youtube.DownloadService) *YoutubeHandler {
	return &YoutubeHandler{
		searchSvc:   searchSvc,
		downloadSvc: downloadSvc,
	}
}

// Search handles GET /search?title=query
func (h *YoutubeHandler) Search(c *gin.Context) {
	query := c.Query("title")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title parameter required"})
		return
	}

	results, err := h.searchSvc.Search(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// Download handles POST /download
func (h *YoutubeHandler) Download(c *gin.Context) {
	var req youtube.DownloadRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	task, err := h.downloadSvc.StartDownload(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, task)
}

// TaskStatus handles GET /task-status/:id
func (h *YoutubeHandler) TaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	task, err := h.downloadSvc.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// Playlist handles GET /playlist?song=title
func (h *YoutubeHandler) Playlist(c *gin.Context) {
	// This depends on your HLS implementation
	// For now, just return the file path
	songTitle := c.Query("song")
	c.JSON(http.StatusOK, gin.H{"song": songTitle})
}
```

---

## 10. Main Server Setup

**File: `cmd/server/main.go` (additions)**

```go
package main

import (
	"log"
	"github.com/yourusername/yourproject/internal/config"
	"github.com/yourusername/yourproject/internal/services/youtube"
	"github.com/yourusername/yourproject/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.NewConfig()

	// Initialize services
	searchSvc := youtube.NewSearchService(cfg.YtDlpPath)
	downloadSvc := youtube.NewDownloadService(
		cfg.StoragePath,
		cfg.TempPath,
		cfg.YtDlpPath,
		cfg.DownloadTimeout,
		500, // max 500MB
	)

	// Initialize handlers
	youtubeHandler := handlers.NewYoutubeHandler(searchSvc, downloadSvc)

	// Setup Gin router
	router := gin.Default()

	// YouTube routes
	router.GET("/search", youtubeHandler.Search)
	router.POST("/download", youtubeHandler.Download)
	router.GET("/task-status/:id", youtubeHandler.TaskStatus)
	router.GET("/playlist", youtubeHandler.Playlist)

	// Start server
	log.Printf("Starting server on %s:%s", cfg.APIHost, cfg.APIPort)
	if err := router.Run(cfg.APIHost + ":" + cfg.APIPort); err != nil {
		log.Fatal(err)
	}
}
```

---

## 11. Installation & Setup

### Prerequisites
```bash
# Install yt-dlp
brew install yt-dlp        # macOS
# or
pip install yt-dlp         # All platforms

# Verify installation
yt-dlp --version
```

### Go Dependencies
```bash
go get github.com/gin-gonic/gin
go get github.com/99designs/gqlgen
go get github.com/joho/godotenv

go mod tidy
```

### Create directories
```bash
mkdir -p media/downloads
mkdir -p media/temp
```

---

## 12. Usage Examples

### GraphQL Query (Search)
```graphql
query {
  searchYoutube(query: "harry styles sweet creature") {
    id
    title
    channel
    thumbnail
    viewCount
    durationString
    url
  }
}
```

### GraphQL Mutation (Download)
```graphql
mutation {
  startDownload(input: {
    title: "Harry Styles - Sweet Creature"
    youtubeUrl: "https://www.youtube.com/watch?v=8uD6s-X3590"
  }) {
    id
    status
    progress
  }
}
```

### REST API Examples
```bash
# Search
curl "http://localhost:3001/search?title=harry%20styles%20sweet%20creature"

# Download
curl -X POST http://localhost:3001/download \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Harry Styles - Sweet Creature",
    "youtube_url": "https://www.youtube.com/watch?v=8uD6s-X3590"
  }'

# Check status
curl http://localhost:3001/task-status/task_1234567890
```

---

## 13. Production Considerations

1. **Database**: Replace in-memory `tasks` map with PostgreSQL/MongoDB
2. **Async Queue**: Use Redis/RabbitMQ for download tasks instead of goroutines
3. **File Storage**: Use S3/GCS for cloud storage
4. **Caching**: Cache search results with Redis
5. **Rate Limiting**: Add request rate limiting
6. **Logging**: Implement structured logging (zerolog/logrus)
7. **Error Handling**: Add proper error recovery and rollback
8. **Testing**: Add unit and integration tests

---

## 14. Troubleshooting

| Issue | Solution |
|-------|----------|
| `yt-dlp not found` | Install yt-dlp and verify PATH |
| `Download hangs` | Check DOWNLOAD_TIMEOUT in config |
| `File permissions` | Ensure write access to STORAGE_PATH |
| `Rate limited by YouTube` | Add delays between requests or use proxies |
| `Large files` | Adjust MAX_DOWNLOAD_SIZE in config |

