package youtube

import (
	"context"
	"testing"
	"time"
)

func TestParseProgressPercent(t *testing.T) {
	tests := []struct {
		input       string
		wantPercent int
		wantOk      bool
	}{
		{"download-progress:  0.4%", 0, true},
		{"download-progress: 12.6%", 12, true},
		{"download-progress: 51.0%", 51, true},
		{"download-progress:100.0%", 100, true},
		{"download-progress: 105.0%", 100, true},
		{"[youtube] Extracting URL: ...", 0, false},
		{"[download] Destination: ...", 0, false},
		{"download-progress: invalid%", 0, false},
		{"", 0, false},
	}

	for _, tt := range tests {
		gotPercent, gotOk := parseProgressPercent(tt.input)
		if gotOk != tt.wantOk {
			t.Errorf("parseProgressPercent(%q) ok = %v, want %v", tt.input, gotOk, tt.wantOk)
		}
		if gotPercent != tt.wantPercent {
			t.Errorf("parseProgressPercent(%q) percent = %v, want %v", tt.input, gotPercent, tt.wantPercent)
		}
	}
}

func TestDownloadServiceSubscribe(t *testing.T) {
	svc := &DownloadService{
		tasks:     make(map[string]*DownloadTask),
		listeners: make(map[string][]chan *DownloadTask),
	}

	taskID := "test-task-123"
	svc.tasks[taskID] = &DownloadTask{
		ID:       taskID,
		Status:   "pending",
		Progress: 0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := svc.Subscribe(ctx, taskID)
	if err != nil {
		t.Fatalf("unexpected error on Subscribe: %v", err)
	}

	// Should receive initial task snapshot immediately
	select {
	case task := <-ch:
		if task.Progress != 0 || task.Status != "pending" {
			t.Errorf("expected initial task state (0, pending), got (%d, %s)", task.Progress, task.Status)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for initial task snapshot")
	}

	// Trigger update
	svc.updateTask(taskID, "downloading", 45, "")

	select {
	case task := <-ch:
		if task.Progress != 45 || task.Status != "downloading" {
			t.Errorf("expected updated task state (45, downloading), got (%d, %s)", task.Progress, task.Status)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for updated task snapshot")
	}

	// Test cleanup on context cancel
	cancel()
	select {
	case _, ok := <-ch:
		if ok {
			// drained buffered, next should close
			<-ch
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for channel close after ctx cancel")
	}
}
