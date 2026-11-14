package resolvers

import (
	"math"
	"strings"
	"testing"
)

func TestClampToInt32(t *testing.T) {
	for value, want := range map[int64]int{0: 0, 183052050: 183052050, 2380307139: math.MaxInt32} {
		if got := clampToInt32(value); got != want {
			t.Errorf("clampToInt32(%d) = %d, want %d", value, got, want)
		}
	}
}

func TestCleanPlaylistTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "keeps a normal title", input: "Road Trip", want: "Road Trip"},
		{name: "trims spaces", input: "  Road Trip  ", want: "Road Trip"},
		{name: "rejects empty", input: "", wantErr: true},
		{name: "rejects spaces only", input: "   ", wantErr: true},
		{name: "allows 100 characters", input: strings.Repeat("a", 100), want: strings.Repeat("a", 100)},
		{name: "rejects 101 characters", input: strings.Repeat("a", 101), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cleanPlaylistTitle(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("cleanPlaylistTitle(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("cleanPlaylistTitle(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
