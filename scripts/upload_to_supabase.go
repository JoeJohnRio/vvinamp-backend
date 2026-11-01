package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	supabaseURL = "https://bumcqqhbtmmrkeiwlsyy.supabase.co"
	bucketName  = "media"
	storageDir  = "/Users/joel/Programming/jerman/Storage"
)

var supabaseKey = os.Getenv("SUPABASE_KEY")

func getMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".mp3":
		return "audio/mpeg"
	case ".m3u8":
		return "application/x-mpegURL"
	case ".ts":
		return "video/MP2T"
	default:
		return "application/octet-stream"
	}
}

func uploadFile(relPath string, fullPath string) error {
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	mimeType := getMimeType(fullPath)
	targetURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, relPath)

	req, err := http.NewRequest("POST", targetURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create req: %w", err)
	}

	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", mimeType)
	req.Header.Set("x-upsert", "true") // overwrite if already exists

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do req: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("upload failed status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Uploaded [%s] -> %s\n", mimeType, relPath)
	return nil
}

func main() {
	foldersToUpload := []string{"artist", "Mobile"}

	for _, folder := range foldersToUpload {
		folderPath := filepath.Join(storageDir, folder)
		err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(info.Name(), ".DS_Store") {
				return nil
			}

			relPath, err := filepath.Rel(storageDir, path)
			if err != nil {
				return err
			}

			// Normalize slashes for URL
			relPath = filepath.ToSlash(relPath)

			return uploadFile(relPath, path)
		})

		if err != nil {
			fmt.Printf("Error walking %s: %v\n", folder, err)
		}
	}

	fmt.Println("Upload completed successfully!")
}
