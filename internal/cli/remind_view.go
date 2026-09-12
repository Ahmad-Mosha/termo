package cli

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/Ahmad-Mosha/termo/internal/remind"
	"github.com/Ahmad-Mosha/termo/internal/ui"
	"github.com/charmbracelet/x/ansi"
)

// maxMessage is how much of a message fits in the list before it's cut.
const maxMessage = 48

// printDetails shows a reminder's message and timing under a status line:
//
//	Standup
//	every weekday at 09:00 · next Mon 09:00 · in 2d 18h
func printDetails(w io.Writer, r remind.Reminder, now time.Time) {
	var parts []string
	switch {
	case r.Done:
		parts = append(parts, "fired "+ui.When(r.FiredAt, now))
	case r.Repeat != nil:
		parts = append(parts, schedule(r.Repeat), "next "+ui.When(r.Due, now), ui.Until(r.Due, now))
	default:
		parts = append(parts, ui.When(r.Due, now), ui.Until(r.Due, now))
	}
	fmt.Fprintln(w, "  "+ui.Bold.Render(r.Message))
	fmt.Fprintln(w, "  "+ui.Faint.Render(strings.Join(parts, " · ")))
}

// printFired shows a reminder that went off:
//
//	⏰ Check the oven · today 14:32 #4
func printFired(w io.Writer, f remind.Fired, now time.Time) {
	fmt.Fprintf(w, "⏰ %s %s %s\n", ui.Bold.Render(f.Message), ui.Faint.Render("· "+ui.When(f.At, now)), idLabel(f.ID))
}

func printTable(w io.Writer, reminders []remind.Reminder, now time.Time) {
	rows := make([][]string, len(reminders))
	for i, r := range reminders {
		rows[i] = reminderRow(r, now)
	}
	fmt.Fprintln(w)
	ui.Table(w, []string{"ID", "MESSAGE", "WHEN", "IN", "REPEATS"}, rows)
}

func reminderRow(r remind.Reminder, now time.Time) []string {
	msg := ansi.Truncate(r.Message, maxMessage, "…")
	repeats := ""
	if r.Repeat != nil {
		repeats = schedule(r.Repeat)
	}
	if r.Done {
		return []string{ui.Faint.Render(fmt.Sprintf("#%d", r.ID)), ui.Faint.Render(msg),
			ui.Faint.Render(ui.When(r.FiredAt, now)), ui.Faint.Render("done"), ui.Faint.Render(repeats)}
	}
	in := ui.Duration(r.Due.Sub(now))
	if !r.Due.After(now) {
		in = ui.Red.Render("overdue")
	}
	return []string{idLabel(r.ID), msg, ui.When(r.Due, now), in, ui.Faint.Render(repeats)}
}

func idLabel(id int) string {
	return ui.Accent.Render(fmt.Sprintf("#%d", id))
}

// schedule describes a repeat schedule, like "every 2h" or
// "every weekday at 09:00".
func schedule(s *remind.Repeat) string {
	if s.Every > 0 {
		names := map[time.Duration]string{
			time.Minute: "minute", time.Hour: "hour", 24 * time.Hour: "day", 7 * 24 * time.Hour: "week",
		}
		if name, ok := names[s.Every]; ok {
			return "every " + name
		}
		return "every " + ui.Compact(s.Every)
	}
	return fmt.Sprintf("every %s at %02d:%02d", dayList(s.Days), s.At/60, s.At%60)
}

// dayList names a set of weekdays: "day", "weekday", "weekend" or
// "Mon, Wed, Fri".
func dayList(days []time.Weekday) string {
	is := func(want ...time.Weekday) bool {
		return len(days) == len(want) && !slices.ContainsFunc(want, func(d time.Weekday) bool {
			return !slices.Contains(days, d)
		})
	}
	switch {
	case len(days) == 7:
		return "day"
	case is(time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday):
		return "weekday"
	case is(time.Saturday, time.Sunday):
		return "weekend"
	}
	sorted := slices.Clone(days)
	mondayFirst := func(d time.Weekday) int { return (int(d) + 6) % 7 }
	slices.SortFunc(sorted, func(a, b time.Weekday) int { return mondayFirst(a) - mondayFirst(b) })
	names := make([]string, len(sorted))
	for i, d := range sorted {
		names[i] = d.String()[:3]
	}
	return strings.Join(names, ", ")
}
