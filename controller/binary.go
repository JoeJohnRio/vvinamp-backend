package controller

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// resolveBinary finds an executable binary either from an environment variable,
// the system PATH, common platform directories (/opt/homebrew/bin, /usr/local/bin, /usr/bin, ./bin),
// or by auto-downloading yt-dlp if it is missing on Linux/macOS.
func resolveBinary(envVar, name string) string {
	// 1. If explicit env var is set, verify it actually exists or resolves in PATH
	if envVal := os.Getenv(envVar); envVal != "" {
		if path, err := exec.LookPath(envVal); err == nil {
			return path
		}
		if fi, err := os.Stat(envVal); err == nil && !fi.IsDir() {
			return envVal
		}
	}

	// 2. Candidates list covering macOS Homebrew (Apple Silicon & Intel), Linux/Docker, and local ./bin
	candidates := []string{
		name,
		filepath.Join("/opt/homebrew/bin", name),
		filepath.Join("/usr/local/bin", name),
		filepath.Join("/usr/bin", name),
		filepath.Join("./bin", name),
		filepath.Join("bin", name),
	}

	for _, cand := range candidates {
		if path, err := exec.LookPath(cand); err == nil {
			return path
		}
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			return cand
		}
	}

	// 3. Auto-download yt-dlp if missing on Linux or macOS
	if name == "yt-dlp" && (runtime.GOOS == "linux" || runtime.GOOS == "darwin") {
		if downloadedPath, err := downloadYtDlp(); err == nil {
			return downloadedPath
		}
	}

	return name
}

func downloadYtDlp() (string, error) {
	binDir := "./bin"
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create %s: %w", binDir, err)
	}
	targetPath := filepath.Join(binDir, "yt-dlp")
	if fi, err := os.Stat(targetPath); err == nil && !fi.IsDir() {
		return targetPath, nil
	}

	log.Printf("[vvinamp] yt-dlp not found on system; auto-downloading standalone binary to %s...", targetPath)
	url := "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp"
	if runtime.GOOS == "darwin" {
		url = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos"
	}

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("[vvinamp] Failed to auto-download yt-dlp: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[vvinamp] Failed to auto-download yt-dlp: HTTP %d", resp.StatusCode)
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	tmpFile := targetPath + ".tmp"
	out, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(tmpFile)
		return "", fmt.Errorf("failed to write yt-dlp binary: %w", err)
	}
	out.Close()

	if err := os.Rename(tmpFile, targetPath); err != nil {
		return "", fmt.Errorf("failed to finalize yt-dlp binary: %w", err)
	}
	log.Printf("[vvinamp] Successfully installed yt-dlp to %s", targetPath)
	return targetPath, nil
}
