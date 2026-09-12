package ui

import (
	"fmt"
	"strings"
	"time"
)

// When formats t for people, relative to now: "today 14:32",
// "tomorrow 09:00", "Mon 09:00", "Sep 20 09:00" or "2027-01-05 09:00".
func When(t, now time.Time) string {
	t = t.In(now.Location())
	switch days := daysBetween(now, t); {
	case days == 0:
		return "today " + t.Format("15:04")
	case days == 1:
		return "tomorrow " + t.Format("15:04")
	case days == -1:
		return "yesterday " + t.Format("15:04")
	case days > 1 && days < 7:
		return t.Format("Mon 15:04")
	case t.Year() == now.Year():
		return t.Format("Jan 2 15:04")
	default:
		return t.Format("2006-01-02 15:04")
	}
}

// Until formats how far t is from now: "in 20m", "in 2d 3h" or "5m ago".
func Until(t, now time.Time) string {
	switch d := t.Sub(now); {
	case d >= time.Second:
		return "in " + Duration(d)
	case d > -time.Second:
		return "now"
	default:
		return Duration(-d) + " ago"
	}
}

// Duration formats d with its two largest units, rounded: "2d 3h", "1h 5m",
// "45s".
func Duration(d time.Duration) string {
	switch {
	case d >= 24*time.Hour:
		d = d.Round(time.Hour)
	case d >= time.Hour:
		d = d.Round(time.Minute)
	default:
		d = d.Round(time.Second)
	}
	if d < time.Second {
		return "0s"
	}
	return strings.Join(units(d), " ")
}

// Compact formats d exactly without spaces: "1h30m", "2d".
func Compact(d time.Duration) string {
	return strings.Join(units(d), "")
}

// units splits d into its non-zero days, hours, minutes and seconds.
func units(d time.Duration) []string {
	var out []string
	for _, u := range []struct {
		size time.Duration
		name string
	}{{24 * time.Hour, "d"}, {time.Hour, "h"}, {time.Minute, "m"}, {time.Second, "s"}} {
		if n := d / u.size; n > 0 {
			out = append(out, fmt.Sprintf("%d%s", n, u.name))
			d -= n * u.size
		}
	}
	return out
}

// daysBetween counts calendar days from a to b.
func daysBetween(a, b time.Time) int {
	midnight := func(t time.Time) time.Time {
		y, m, d := t.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	}
	return int(midnight(b).Sub(midnight(a)).Hours() / 24)
}
