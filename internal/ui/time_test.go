package ui

import (
	"testing"
	"time"
)

func TestWhen(t *testing.T) {
	now := time.Date(2026, 9, 16, 22, 0, 0, 0, time.UTC) // a Wednesday
	tests := map[time.Time]string{
		now.Add(30 * time.Minute):                     "today 22:30",
		now.Add(3 * time.Hour):                        "tomorrow 01:00",
		now.Add(-23 * time.Hour):                      "yesterday 23:00",
		now.AddDate(0, 0, 5):                          "Mon 22:00",
		time.Date(2026, 12, 25, 9, 0, 0, 0, time.UTC): "Dec 25 09:00",
		time.Date(2027, 1, 5, 9, 0, 0, 0, time.UTC):   "2027-01-05 09:00",
	}
	for in, want := range tests {
		if got := When(in, now); got != want {
			t.Errorf("When(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestUntil(t *testing.T) {
	now := time.Date(2026, 9, 16, 10, 0, 0, 0, time.UTC)
	tests := map[time.Duration]string{
		20 * time.Minute: "in 20m",
		90 * time.Second: "in 1m 30s",
		time.Hour + 5*time.Minute + 20*time.Second:     "in 1h 5m",
		2*24*time.Hour + 3*time.Hour + 40*time.Minute:  "in 2d 4h",
		23*time.Hour + 59*time.Minute + 50*time.Second: "in 1d",
		0:                "now",
		-5 * time.Minute: "5m ago",
	}
	for in, want := range tests {
		if got := Until(now.Add(in), now); got != want {
			t.Errorf("Until(now+%v) = %q, want %q", in, got, want)
		}
	}
}
