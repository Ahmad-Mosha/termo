package remind

import (
	"reflect"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	good := map[string]time.Duration{
		"20m":               20 * time.Minute,
		"1h30m":             90 * time.Minute,
		"2d":                48 * time.Hour,
		"1w":                7 * 24 * time.Hour,
		"45s":               45 * time.Second,
		"10 mins":           10 * time.Minute,
		"1 hour 30 minutes": 90 * time.Minute,
	}
	for in, want := range good {
		if got, err := ParseDuration(in); err != nil || got != want {
			t.Errorf("ParseDuration(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, in := range []string{"", "20", "0m", "soon", "1 month", "20m later", "1234567m"} {
		if got, err := ParseDuration(in); err == nil {
			t.Errorf("ParseDuration(%q) = %v; want an error", in, got)
		}
	}
}

func TestParseTime(t *testing.T) {
	now := wed(10, 0)
	thu := func(h, m int) time.Time { return wed(h, m).AddDate(0, 0, 1) }
	good := map[string]time.Time{
		"15:30":            wed(15, 30),
		"noon":             wed(12, 0),
		"9am":              thu(9, 0), // already passed today
		"9 AM":             thu(9, 0),
		"tomorrow":         thu(9, 0),
		"tomorrow 8:15pm":  thu(20, 15),
		"8:15pm tomorrow":  thu(20, 15),
		"wed 11am":         wed(11, 0),
		"wed 9am":          wed(9, 0).AddDate(0, 0, 7),
		"fri 18:00":        wed(18, 0).AddDate(0, 0, 2),
		"2026-09-20":       time.Date(2026, 9, 20, 9, 0, 0, 0, time.UTC),
		"2026-09-20 07:45": time.Date(2026, 9, 20, 7, 45, 0, 0, time.UTC),
	}
	for in, want := range good {
		if got, err := ParseTime(in, now); err != nil || !got.Equal(want) {
			t.Errorf("ParseTime(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	bad := []string{"", "9", "25:00", "13pm", "today 9am", "2020-01-01 10:00", "someday", "next friday", "9am 10am"}
	for _, in := range bad {
		if got, err := ParseTime(in, now); err == nil {
			t.Errorf("ParseTime(%q) = %v; want an error", in, got)
		}
	}
}

func TestParseRepeat(t *testing.T) {
	good := map[string]Repeat{
		"2h":                 {Every: 2 * time.Hour},
		"hour":               {Every: time.Hour},
		"30 minutes":         {Every: 30 * time.Minute},
		"day 09:00":          {Days: allDays, At: 9 * 60},
		"weekday at 9am":     {Days: workDays, At: 9 * 60},
		"weekends noon":      {Days: weekend, At: 12 * 60},
		"mon,wed,fri 6:30pm": {Days: []time.Weekday{time.Monday, time.Wednesday, time.Friday}, At: 18*60 + 30},
	}
	for in, want := range good {
		if got, err := ParseRepeat(in); err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("ParseRepeat(%q) = %+v, %v; want %+v", in, got, err, want)
		}
	}
	for _, in := range []string{"30s", "weekday", "day 9", "funday 9am", "mon,xyz 9am"} {
		if got, err := ParseRepeat(in); err == nil {
			t.Errorf("ParseRepeat(%q) = %+v; want an error", in, got)
		}
	}
}
