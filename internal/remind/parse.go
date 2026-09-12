package remind

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// defaultClock is the time used when only a day is given, like "tomorrow".
const defaultClock = 9 * 60

var (
	durationPart = regexp.MustCompile(`(\d{1,6})\s*(weeks?|w|days?|d|hours?|hrs?|h|minutes?|mins?|m|seconds?|secs?|s)`)
	clockPattern = regexp.MustCompile(`^(\d{1,2})(?::(\d{2}))?(am|pm)?$`)
	spacedAmPm   = regexp.MustCompile(`(\d)\s+(am|pm)\b`)
)

var units = map[byte]time.Duration{
	'w': 7 * 24 * time.Hour,
	'd': 24 * time.Hour,
	'h': time.Hour,
	'm': time.Minute,
	's': time.Second,
}

var dayNames = map[string]time.Weekday{
	"sun": time.Sunday, "sunday": time.Sunday,
	"mon": time.Monday, "monday": time.Monday,
	"tue": time.Tuesday, "tues": time.Tuesday, "tuesday": time.Tuesday,
	"wed": time.Wednesday, "wednesday": time.Wednesday,
	"thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday, "thursday": time.Thursday,
	"fri": time.Friday, "friday": time.Friday,
	"sat": time.Saturday, "saturday": time.Saturday,
}

var (
	allDays  = []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday}
	workDays = allDays[:5]
	weekend  = allDays[5:]
)

var dayGroups = map[string][]time.Weekday{
	"day": allDays, "days": allDays, "daily": allDays, "everyday": allDays,
	"weekday": workDays, "weekdays": workDays,
	"weekend": weekend, "weekends": weekend,
}

// ParseDuration reads durations like "20m", "1h30m", "2d" or "1 hour 30 minutes".
func ParseDuration(s string) (time.Duration, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	parts := durationPart.FindAllStringSubmatch(s, -1)
	if len(parts) == 0 || strings.TrimSpace(durationPart.ReplaceAllString(s, "")) != "" {
		return 0, fmt.Errorf("%q isn't a duration (try 20m, 1h30m or 2d)", s)
	}
	var d time.Duration
	for _, p := range parts {
		n, _ := strconv.Atoi(p[1])
		d += time.Duration(n) * units[p[2][0]]
	}
	if d <= 0 {
		return 0, errors.New("the duration must be longer than zero")
	}
	return d, nil
}

// ParseTime reads a moment like "15:30", "9am", "tomorrow 9am", "fri 18:00"
// or "2026-09-20 09:00". A clock time alone means the next time that clock
// comes around; a day alone means 09:00 that day.
func ParseTime(s string, now time.Time) (time.Time, error) {
	bad := fmt.Errorf("%q isn't a time (try 15:30, 9am, tomorrow 9am, fri 18:00 or 2026-09-20 09:00)", s)
	words := strings.Fields(spacedAmPm.ReplaceAllString(strings.ToLower(s), "$1$2"))
	clock, day := defaultClock, ""
	switch len(words) {
	case 1:
		if c, ok := parseClock(words[0]); ok {
			clock = c
		} else {
			day = words[0]
		}
	case 2:
		day = words[0]
		c, ok := parseClock(words[1])
		if !ok { // also accept "9am tomorrow"
			if c, ok = parseClock(words[0]); !ok {
				return time.Time{}, bad
			}
			day = words[1]
		}
		clock = c
	default:
		return time.Time{}, bad
	}

	var t time.Time
	switch wd, isWeekday := dayNames[day]; {
	case day == "":
		if t = atClock(now, clock); !t.After(now) {
			t = atClock(now.AddDate(0, 0, 1), clock)
		}
	case day == "today":
		t = atClock(now, clock)
	case day == "tomorrow":
		t = atClock(now.AddDate(0, 0, 1), clock)
	case isWeekday:
		for i := range 8 {
			if t = atClock(now.AddDate(0, 0, i), clock); t.Weekday() == wd && t.After(now) {
				break
			}
		}
	default:
		date, err := time.ParseInLocation(time.DateOnly, day, now.Location())
		if err != nil {
			return time.Time{}, bad
		}
		t = atClock(date, clock)
	}
	if !t.After(now) {
		return time.Time{}, fmt.Errorf("%q is in the past", s)
	}
	return t, nil
}

// ParseRepeat reads a schedule: an interval like "2h" or "hour", or days
// and a time like "day 09:00", "weekday at 9am" or "mon,wed,fri 18:00".
func ParseRepeat(s string) (Repeat, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	d, err := ParseDuration(s)
	if err != nil && !strings.ContainsAny(s, "0123456789") {
		d, err = ParseDuration("1 " + s) // a bare unit: "hour" means "1 hour"
	}
	if err == nil {
		if d < time.Minute {
			return Repeat{}, errors.New("reminders can repeat at most once a minute")
		}
		return Repeat{Every: d}, nil
	}

	bad := fmt.Errorf("%q isn't a schedule (try 2h, day 09:00, weekday 9am or mon,wed,fri 18:00)", s)
	words := strings.Fields(spacedAmPm.ReplaceAllString(s, "$1$2"))
	if len(words) == 3 && words[1] == "at" {
		words = []string{words[0], words[2]}
	}
	if len(words) != 2 {
		return Repeat{}, bad
	}
	days, ok := parseDays(words[0])
	clock, ok2 := parseClock(words[1])
	if !ok || !ok2 {
		return Repeat{}, bad
	}
	return Repeat{Days: days, At: clock}, nil
}

// parseClock reads "9am", "9:30pm", "21:30", "noon" or "midnight" as minutes
// after midnight. A bare "9" is rejected because am or pm is unclear.
func parseClock(s string) (int, bool) {
	switch s {
	case "noon":
		return 12 * 60, true
	case "midnight":
		return 0, true
	}
	m := clockPattern.FindStringSubmatch(s)
	if m == nil || (m[2] == "" && m[3] == "") {
		return 0, false
	}
	h, _ := strconv.Atoi(m[1])
	mins, _ := strconv.Atoi(m[2])
	if m[3] != "" {
		if h < 1 || h > 12 {
			return 0, false
		}
		h %= 12
		if m[3] == "pm" {
			h += 12
		}
	}
	if h > 23 || mins > 59 {
		return 0, false
	}
	return h*60 + mins, true
}

// parseDays reads "day", "weekday", "weekend" or names like "mon,wed,fri".
func parseDays(s string) ([]time.Weekday, bool) {
	if days, ok := dayGroups[s]; ok {
		return slices.Clone(days), true
	}
	var days []time.Weekday
	for _, name := range strings.Split(s, ",") {
		d, ok := dayNames[name]
		if !ok {
			return nil, false
		}
		if !slices.Contains(days, d) {
			days = append(days, d)
		}
	}
	return days, true
}
