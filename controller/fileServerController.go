package controller

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// FileServerControllerTarget is parameter object for getting all FileServerController's dependency
type FileServerControllerTarget struct {
	fx.In
	StoragePath string `name:"storage_path"`
}

// FileServerController handles serving static files from storage directory
type FileServerController struct {
	storagePath string
}

// NewFileServerController is a constructor for FileServerController
func NewFileServerController(target FileServerControllerTarget) Result {
	storagePath := target.StoragePath
	if storagePath == "" {
		// Default path if not configured
		storagePath = "S:\\Project German\\Backend\\Storage"
	}

	return Result{
		Controller: &FileServerController{
			storagePath: storagePath,
		},
	}
}

// Register is function to register all controller's endpoint handler
func (f *FileServerController) Register(r *gin.Engine) {
	// Serve static files from the storage directory
	// This will serve files at routes like /storage/images/, /storage/tracks/, etc.
	r.Static("/storage", f.storagePath)

	// You can also create specific routes for different file types
	// These will serve from specific subdirectories in storage

	// r.Static("/hls", filepath.Join(f.storagePath, "hls"))
	// r.Static("/images", filepath.Join(f.storagePath, "images"))
	// r.Static("/tracks", filepath.Join(f.storagePath, "tracks"))
	// r.Static("/files", filepath.Join(f.storagePath, "files"))
	r.Static("/artist", filepath.Join(f.storagePath, "artist"))
	// r.Static("/vvinamp", filepath.Join(f.storagePath, "/vvinamp/artist/michael_jackson")) //this works
	// r.Static("/vvinamp", filepath.Join(f.storagePath, "/vvinamp/artist"))//this doesnt works
}

func (f *FileServerController) ConvertToHLS(ctx context.Context, mp3RelativePath, trackId string) (string, error) {
	mp3Path := filepath.Join(f.storagePath, mp3RelativePath) // e.g. "mp3/track123.mp3"

	hlsDir := filepath.Join(f.storagePath, "hls", trackId)
	if err := os.MkdirAll(hlsDir, 0755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", hlsDir, err)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", mp3Path,
		"-c:a", "aac", "-b:a", "128k",
		"-f", "hls",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(hlsDir, "seg_%03d.ts"),
		filepath.Join(hlsDir, "playlist.m3u8"))

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %v\noutput: %s", err, output)
	}

	// Return public URL (adjust baseURL from config/resolver)
	return fmt.Sprintf("/hls/%s/playlist.m3u8", trackId), nil
}

// Bonus: Check if HLS exists (avoid reconvert)
func (f *FileServerController) HLSExists(trackId string) bool {
	return fileExists(filepath.Join(f.storagePath, "hls", trackId, "playlist.m3u8"))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
