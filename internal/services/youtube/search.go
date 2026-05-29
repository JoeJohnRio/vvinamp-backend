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
	return &SearchService{ytDlpPath: ytDlpPath}
}

func (s *SearchService) Search(ctx context.Context, query string) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	searchURL := fmt.Sprintf("ytsearch5:%s", query)
	cmd := exec.CommandContext(ctx, s.ytDlpPath,
		"--dump-json",
		"--skip-download",
		searchURL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp search failed: %w: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return []SearchResult{}, nil
	}

	lines := strings.Split(output, "\n")
	results := make([]SearchResult, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var result SearchResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			return nil, fmt.Errorf("failed to parse yt-dlp search result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}
