package supabase

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestContentTypeByExt(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{".jpg", "image/jpeg"},
		{".JPEG", "image/jpeg"},
		{".png", "image/png"},
		{".webp", "image/webp"},
		{".mp3", "audio/mpeg"},
		{".unknown", "application/octet-stream"},
	}

	for _, tt := range tests {
		got := ContentTypeByExt(tt.ext)
		if got != tt.expected {
			t.Errorf("ContentTypeByExt(%q) = %q; want %q", tt.ext, got, tt.expected)
		}
	}
}

func TestStorageClientDefaults(t *testing.T) {
	client := NewStorageClient()
	if client.BaseURL == "" {
		t.Errorf("expected default BaseURL to be non-empty")
	}
	if client.Bucket != "media" {
		t.Errorf("expected default Bucket to be 'media', got %q", client.Bucket)
	}
	if strings.HasSuffix(client.BaseURL, "/") {
		t.Errorf("expected BaseURL to not have trailing slash, got %q", client.BaseURL)
	}
}

func TestLiveStorageUploadAndDelete(t *testing.T) {
	key := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if key == "" {
		t.Skip("SUPABASE_SERVICE_ROLE_KEY not configured, skipping live storage test")
	}

	client := NewStorageClient()
	testPath := "Mobile/Playlist/test-unit/playlist.jpg"
	content := []byte("fake-jpeg-data")

	publicURL, err := client.Upload(testPath, bytes.NewReader(content), "image/jpeg")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if !strings.Contains(publicURL, "https://bumcqqhbtmmrkeiwlsyy.supabase.co/storage/v1/object/public/media/Mobile/Playlist/test-unit/playlist.jpg") {
		t.Errorf("unexpected public URL: %s", publicURL)
	}

	resp, err := http.Get(publicURL)
	if err != nil {
		t.Fatalf("failed to fetch uploaded file: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != string(content) {
		t.Errorf("expected content %q, got %q", string(content), string(body))
	}

	if err := client.Delete(testPath); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
