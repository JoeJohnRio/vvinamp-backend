package youtube

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type DownloadService struct {
	storagePath string
	ytDlpPath   string
	timeoutSec  int
	maxSizeMB   int64
	tasks       map[string]*DownloadTask
	tasksMu     sync.RWMutex
	listeners   map[string][]chan *DownloadTask
	listenersMu sync.RWMutex
}

type DownloadResult struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	YoutubeURL string `json:"youtube_url"`
	FilePath   string `json:"file_path"`
}

func NewDownloadService(storagePath, ytDlpPath string, timeoutSec int, maxSizeMB int64) *DownloadService {
	if storagePath == "" {
		storagePath = "./media/downloads"
	}
	storagePath = filepath.Join(storagePath, "youtube")

	if ytDlpPath == "" {
		ytDlpPath = "yt-dlp"
	}

	if timeoutSec <= 0 {
		timeoutSec = 600
	}

	if maxSizeMB <= 0 {
		maxSizeMB = 500
	}

	if err := os.MkdirAll(storagePath, 0o755); err != nil {
		fallback := filepath.Join(os.TempDir(), "vvinamp", "youtube")
		if errFallback := os.MkdirAll(fallback, 0o755); errFallback == nil {
			log.Printf("Warning: cannot use storage path %s (%v), using fallback: %s", storagePath, err, fallback)
			storagePath = fallback
		} else {
			log.Printf("Warning: failed to create download storage folder: %v", err)
		}
	}

	return &DownloadService{
		storagePath: storagePath,
		ytDlpPath:   ytDlpPath,
		timeoutSec:  timeoutSec,
		maxSizeMB:   maxSizeMB,
		tasks:       make(map[string]*DownloadTask),
		listeners:   make(map[string][]chan *DownloadTask),
	}
}

func (d *DownloadService) StartDownload(ctx context.Context, req DownloadRequest) (*DownloadTask, error) {
	if strings.TrimSpace(req.YoutubeURL) == "" {
		return nil, fmt.Errorf("youtube URL is required")
	}

	task := &DownloadTask{
		ID:         uuid.NewString(),
		YoutubeURL: req.YoutubeURL,
		Status:     "pending",
		Progress:   0,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	d.tasksMu.Lock()
	d.tasks[task.ID] = task
	d.tasksMu.Unlock()

	go d.executeDownload(context.Background(), task)
	return task, nil
}

func (d *DownloadService) GetTask(taskID string) (*DownloadTask, error) {
	d.tasksMu.RLock()
	defer d.tasksMu.RUnlock()

	task, ok := d.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("download task not found: %s", taskID)
	}

	return task, nil
}

func (d *DownloadService) ListTasks() []*DownloadTask {
	d.tasksMu.RLock()
	defer d.tasksMu.RUnlock()

	tasks := make([]*DownloadTask, 0, len(d.tasks))
	for _, task := range d.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// parseProgressPercent parses the download percentage from a yt-dlp --progress-template output line.
func parseProgressPercent(line string) (int, bool) {
	const prefix = "download-progress:"
	if !strings.HasPrefix(line, prefix) {
		return 0, false
	}
	raw := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	raw = strings.TrimSuffix(raw, "%")
	val, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, false
	}
	p := int(val)
	if p < 0 {
		p = 0
	} else if p > 100 {
		p = 100
	}
	return p, true
}

func (d *DownloadService) runYtDlpDownload(ctx context.Context, outputPath, youtubeURL string, onProgress func(int)) error {
	cmd := exec.CommandContext(ctx, d.ytDlpPath,
		"-4",
		"--extractor-args", "youtube:player_client=android,web",
		"--no-playlist",
		"--extract-audio",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"--max-filesize", fmt.Sprintf("%dM", d.maxSizeMB),
		"--newline",
		"--progress-template", "download-progress:%(progress._percent_str)s",
		"--output", outputPath,
		youtubeURL,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start yt-dlp: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	lastReported := -1
	for scanner.Scan() {
		line := scanner.Text()
		if p, ok := parseProgressPercent(line); ok {
			if p != lastReported {
				lastReported = p
				if onProgress != nil {
					onProgress(p)
				}
			}
		}
	}
	if lastReported >= 0 {
		fmt.Println()
	}

	if err := cmd.Wait(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if ctx.Err() == context.DeadlineExceeded {
			errMsg = "download timeout"
		}
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Errorf("yt-dlp download failed: %w: %s", err, errMsg)
	}

	return nil
}

func (d *DownloadService) executeDownload(ctx context.Context, task *DownloadTask) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(d.timeoutSec)*time.Second)
	defer cancel()

	d.updateTask(task.ID, "initializing", 5, "")

	meta, err := d.fetchMetadata(ctx, task.YoutubeURL)
	if err != nil {
		d.updateTask(task.ID, "failed", 0, err.Error())
		return
	}

	d.updateTask(task.ID, "downloading", 0, "")
	if meta.Title != "" {
		task.Title = meta.Title
	}

	outputTemplate := filepath.Join(d.storagePath, "%(id)s.mp3")
	displayID := task.ID
	if len(displayID) > 8 {
		displayID = displayID[:8]
	}

	err = d.runYtDlpDownload(ctx, outputTemplate, task.YoutubeURL, func(percent int) {
		d.updateTask(task.ID, "downloading", percent, "")
		fmt.Printf("\r[yt-dlp] Downloading [%s]: %3d%%", displayID, percent)
	})
	if err != nil {
		d.updateTask(task.ID, "failed", 0, err.Error())
		return
	}

	d.updateTask(task.ID, "finalizing", 95, "")

	finalPath := filepath.Join(d.storagePath, fmt.Sprintf("%s.mp3", meta.ID))
	if _, err := os.Stat(finalPath); err != nil {
		fallbackPath := filepath.Join(d.storagePath, fmt.Sprintf("%s.m4a", meta.ID))
		if _, err2 := os.Stat(fallbackPath); err2 == nil {
			finalPath = fallbackPath
		}
	}

	d.updateTask(task.ID, "completed", 100, "")
	d.tasksMu.Lock()
	task.FilePath = finalPath
	task.UpdatedAt = time.Now().UTC()
	d.tasksMu.Unlock()
}

func (d *DownloadService) EnsureDownloaded(ctx context.Context, youtubeURL string) (*DownloadResult, error) {
	if strings.TrimSpace(youtubeURL) == "" {
		return nil, fmt.Errorf("youtube url is required")
	}

	meta, err := d.fetchMetadata(ctx, youtubeURL)
	if err != nil {
		return nil, err
	}

	finalPath := filepath.Join(d.storagePath, fmt.Sprintf("%s.mp3", meta.ID))
	if _, err := os.Stat(finalPath); err == nil {
		return &DownloadResult{ID: meta.ID, Title: meta.Title, YoutubeURL: youtubeURL, FilePath: finalPath}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(d.timeoutSec)*time.Second)
	defer cancel()

	err = d.runYtDlpDownload(ctx, finalPath, youtubeURL, func(percent int) {
		fmt.Printf("\r[yt-dlp] Downloading [%s]: %3d%%", meta.ID, percent)
	})
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(finalPath); err != nil {
		return nil, fmt.Errorf("download completed but file missing: %s", finalPath)
	}

	return &DownloadResult{ID: meta.ID, Title: meta.Title, YoutubeURL: youtubeURL, FilePath: finalPath}, nil
}

func (d *DownloadService) fetchMetadata(ctx context.Context, youtubeURL string) (*DownloadTask, error) {
	cmd := exec.CommandContext(ctx, d.ytDlpPath,
		"-4",
		"--extractor-args", "youtube:player_client=android,web",
		"--skip-download",
		"--print-json",
		youtubeURL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %w: %s", err, stderr.String())
	}

	var meta struct {
		Title string `json:"title"`
		ID    string `json:"id"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &meta); err != nil {
		return nil, fmt.Errorf("failed to decode metadata response: %w", err)
	}

	return &DownloadTask{Title: meta.Title, ID: meta.ID}, nil
}

func (d *DownloadService) updateTask(taskID, status string, progress int, errMsg string) {
	d.tasksMu.Lock()
	task, ok := d.tasks[taskID]
	if !ok {
		d.tasksMu.Unlock()
		return
	}

	task.Status = status
	if progress >= 0 {
		task.Progress = progress
	}
	if errMsg != "" {
		task.Error = errMsg
	}
	task.UpdatedAt = time.Now().UTC()
	snapshot := *task
	d.tasksMu.Unlock()

	d.listenersMu.RLock()
	if channels, exists := d.listeners[taskID]; exists {
		for _, ch := range channels {
			select {
			case ch <- &snapshot:
			default:
			}
		}
	}
	d.listenersMu.RUnlock()
}

// Subscribe returns a channel that receives real-time task updates until ctx is canceled or closed.
func (d *DownloadService) Subscribe(ctx context.Context, taskID string) (<-chan *DownloadTask, error) {
	ch := make(chan *DownloadTask, 16)

	d.listenersMu.Lock()
	d.listeners[taskID] = append(d.listeners[taskID], ch)
	d.listenersMu.Unlock()

	// If task already exists, send current snapshot immediately
	d.tasksMu.RLock()
	if task, ok := d.tasks[taskID]; ok {
		snapshot := *task
		select {
		case ch <- &snapshot:
		default:
		}
	}
	d.tasksMu.RUnlock()

	// Clean up when client disconnects or completes
	go func() {
		<-ctx.Done()
		d.listenersMu.Lock()
		defer d.listenersMu.Unlock()
		channels := d.listeners[taskID]
		for i, c := range channels {
			if c == ch {
				d.listeners[taskID] = append(channels[:i], channels[i+1:]...)
				break
			}
		}
		if len(d.listeners[taskID]) == 0 {
			delete(d.listeners, taskID)
		}
		close(ch)
	}()

	return ch, nil
}
