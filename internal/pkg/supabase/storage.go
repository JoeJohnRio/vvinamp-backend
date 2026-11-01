package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// StorageClient communicates with the Supabase Object Storage REST API.
type StorageClient struct {
	BaseURL    string
	Bucket     string
	ServiceKey string
	HTTPClient *http.Client
}

// NewStorageClient initializes client from environment variables.
func NewStorageClient() *StorageClient {
	baseURL := os.Getenv("SUPABASE_URL")
	if baseURL == "" {
		baseURL = "https://bumcqqhbtmmrkeiwlsyy.supabase.co"
	}
	bucket := os.Getenv("SUPABASE_STORAGE_BUCKET")
	if bucket == "" {
		bucket = "media"
	}
	key := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if key == "" {
		key = os.Getenv("SUPABASE_KEY")
	}

	return &StorageClient{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Bucket:     bucket,
		ServiceKey: key,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Upload streams an object directly to Supabase Storage and returns its public CDN URL.
func (s *StorageClient) Upload(storagePath string, reader io.Reader, contentType string) (string, error) {
	if s.ServiceKey == "" {
		return "", fmt.Errorf("SUPABASE_SERVICE_ROLE_KEY is not configured")
	}

	cleanPath := strings.TrimLeft(storagePath, "/")
	endpoint := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.BaseURL, s.Bucket, cleanPath)

	req, err := http.NewRequest(http.MethodPost, endpoint, reader)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.ServiceKey)
	req.Header.Set("apikey", s.ServiceKey)
	// Overwrite existing file to support playlist cover replacement
	req.Header.Set("x-upsert", "true")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("supabase storage upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase storage upload error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s?v=%d", s.BaseURL, s.Bucket, cleanPath, time.Now().Unix()), nil
}

// Delete removes objects by key from Supabase Storage.
func (s *StorageClient) Delete(storagePaths ...string) error {
	if s.ServiceKey == "" || len(storagePaths) == 0 {
		return nil
	}

	endpoint := fmt.Sprintf("%s/storage/v1/object/%s", s.BaseURL, s.Bucket)

	type deletePayload struct {
		Prefixes []string `json:"prefixes"`
	}
	payloadBytes, err := json.Marshal(deletePayload{Prefixes: storagePaths})
	if err != nil {
		return fmt.Errorf("failed to marshal delete payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, endpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.ServiceKey)
	req.Header.Set("apikey", s.ServiceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("supabase storage delete failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase storage delete error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// ContentTypeByExt returns the standard MIME type for media file extensions.
func ContentTypeByExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}
