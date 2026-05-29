package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
}

func NewDownloadService(storagePath, ytDlpPath string, timeoutSec int, maxSizeMB int64) *DownloadService {
	if storagePath == "" {
		storagePath = "./media/downloads"
	}

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
		panic(fmt.Errorf("failed to create download storage folder: %w", err))
	}

	return &DownloadService{
		storagePath: storagePath,
		ytDlpPath:   ytDlpPath,
		timeoutSec:  timeoutSec,
		maxSizeMB:   maxSizeMB,
		tasks:       make(map[string]*DownloadTask),
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

func (d *DownloadService) executeDownload(ctx context.Context, task *DownloadTask) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(d.timeoutSec)*time.Second)
	defer cancel()

	d.updateTask(task.ID, "initializing", 5, "")

	meta, err := d.fetchMetadata(ctx, task.YoutubeURL)
	if err != nil {
		d.updateTask(task.ID, "failed", 0, err.Error())
		return
	}

	d.updateTask(task.ID, "downloading", 15, "")
	if meta.Title != "" {
		d.updateTask(task.ID, "downloading", 15, "")
		task.Title = meta.Title
	}

	outputTemplate := filepath.Join(d.storagePath, "%(id)s.%(ext)s")
	cmd := exec.CommandContext(ctx, d.ytDlpPath,
		"--no-playlist",
		"--extract-audio",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"--max-filesize", fmt.Sprintf("%dM", d.maxSizeMB),
		"--output", outputTemplate,
		task.YoutubeURL,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if ctx.Err() == context.DeadlineExceeded {
			errMsg = "download timeout"
		}
		d.updateTask(task.ID, "failed", 0, strings.TrimSpace(errMsg))
		return
	}

	d.updateTask(task.ID, "finalizing", 90, "")

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

func (d *DownloadService) fetchMetadata(ctx context.Context, youtubeURL string) (*DownloadTask, error) {
	cmd := exec.CommandContext(ctx, d.ytDlpPath,
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
	defer d.tasksMu.Unlock()

	task, ok := d.tasks[taskID]
	if !ok {
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
}
