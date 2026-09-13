package ui

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestBytes(t *testing.T) {
	tests := map[uint64]string{
		0:             "0 B",
		512:           "512 B",
		1500:          "1.5 KB",
		3_400_000:     "3.4 MB",
		1_200_000_000: "1.2 GB",
	}
	for in, want := range tests {
		if got := Bytes(in); got != want {
			t.Errorf("Bytes(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestBar(t *testing.T) {
	tests := []struct {
		percent  float64
		width    int
		wantFull string // the plain, unstyled bar
	}{
		{0, 10, "░░░░░░░░░░"},
		{50, 10, "█████░░░░░"},
		{100, 10, "██████████"},
		{150, 10, "██████████"}, // clamps above 100
		{-10, 10, "░░░░░░░░░░"}, // clamps below 0
	}
	for _, tt := range tests {
		if got := ansi.Strip(Bar(tt.percent, tt.width)); got != tt.wantFull {
			t.Errorf("Bar(%v, %d) = %q, want %q", tt.percent, tt.width, got, tt.wantFull)
		}
	}
}
