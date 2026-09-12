package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/Ahmad-Mosha/termo/internal/remind"
)

func TestSplitArgs(t *testing.T) {
	now := time.Date(2026, 9, 16, 10, 0, 0, 0, time.UTC)
	parseTime := func(s string) (time.Time, error) { return remind.ParseTime(s, now) }

	d, msg, err := splitArgs([]string{"20m", "Check", "the", "oven"}, remind.ParseDuration)
	if err != nil || d != 20*time.Minute || msg != "Check the oven" {
		t.Errorf("unquoted message: got %v %q %v", d, msg, err)
	}
	for _, args := range [][]string{{"tomorrow 9am", "Call mom"}, {"tomorrow", "9am", "Call", "mom"}} {
		at, msg, err := splitArgs(args, parseTime)
		if err != nil || at.Day() != 17 || at.Hour() != 9 || msg != "Call mom" {
			t.Errorf("%q: got %v %q %v", args, at, msg, err)
		}
	}

	// "tomorrow 9am" is a complete time, so the message is missing rather
	// than being "9am".
	if _, _, err := splitArgs([]string{"tomorrow", "9am"}, parseTime); err == nil || !strings.Contains(err.Error(), "message") {
		t.Errorf("missing message: got %v", err)
	}
	if _, _, err := splitArgs([]string{"20mm", "Check the oven"}, remind.ParseDuration); err == nil || !strings.Contains(err.Error(), `"20mm"`) {
		t.Errorf("typo: got %v, want it to name the bad word", err)
	}
}
