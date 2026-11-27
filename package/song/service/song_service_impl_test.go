package song

import (
	"reflect"
	"testing"
)

func TestTrackIDsOnly(t *testing.T) {
	songID := "3f1c2b9e-8a4d-4c1e-9b7a-2d5e6f708192"
	got := trackIDsOnly([]string{songID, "dQw4w9WgXcQ", ""})
	if want := []string{songID}; !reflect.DeepEqual(got, want) {
		t.Fatalf("trackIDsOnly = %v, want %v", got, want)
	}
}
