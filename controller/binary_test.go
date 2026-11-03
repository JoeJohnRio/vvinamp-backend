package controller

import (
	"os"
	"testing"
)

func TestResolveBinaryWithNonExistent(t *testing.T) {
	// A completely bogus binary should safely return its name without panicking
	res := resolveBinary("NON_EXISTENT_VAR", "non_existent_binary_12345")
	if res != "non_existent_binary_12345" {
		t.Errorf("expected fallback to name, got %s", res)
	}
}

func TestResolveBinaryEnvOverride(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "dummy_bin_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	t.Setenv("TEST_BIN_PATH", tmpFile.Name())
	res := resolveBinary("TEST_BIN_PATH", "dummy")
	if res != tmpFile.Name() {
		t.Errorf("expected %s, got %s", tmpFile.Name(), res)
	}
}

func TestResolveBinaryHostYtDlp(t *testing.T) {
	path := resolveBinary("YT_DLP_PATH", "yt-dlp")
	t.Logf("Resolved yt-dlp path: %s", path)
	if path == "" {
		t.Errorf("expected non-empty path for yt-dlp")
	}
}
