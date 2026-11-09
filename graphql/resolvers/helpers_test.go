package resolvers

import (
	"math"
	"testing"
)

func TestClampToInt32(t *testing.T) {
	for value, want := range map[int64]int{0: 0, 183052050: 183052050, 2380307139: math.MaxInt32} {
		if got := clampToInt32(value); got != want {
			t.Errorf("clampToInt32(%d) = %d, want %d", value, got, want)
		}
	}
}
