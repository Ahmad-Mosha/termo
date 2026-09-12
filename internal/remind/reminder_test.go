package remind

import (
	"testing"
	"time"
)

// wed is Wednesday 2026-09-16 at the given clock time.
func wed(h, m int) time.Time { return time.Date(2026, 9, 16, h, m, 0, 0, time.UTC) }

func TestRepeatNext(t *testing.T) {
	tests := []struct {
		name     string
		repeat   Repeat
		due, now time.Time
		want     time.Time
	}{
		{"interval", Repeat{Every: 2 * time.Hour}, wed(10, 0), wed(10, 0), wed(12, 0)},
		{"interval skips missed runs", Repeat{Every: 2 * time.Hour}, wed(10, 0), wed(15, 30), wed(16, 0)},
		{"interval never returns now", Repeat{Every: 2 * time.Hour}, wed(10, 0), wed(16, 0), wed(18, 0)},
		{"later today", Repeat{Days: workDays, At: 9 * 60}, wed(8, 0), wed(8, 0), wed(9, 0)},
		{"tomorrow", Repeat{Days: workDays, At: 9 * 60}, wed(9, 0), wed(9, 0), wed(9, 0).AddDate(0, 0, 1)},
		{"skips the weekend", Repeat{Days: workDays, At: 9 * 60}, wed(9, 0), wed(9, 0).AddDate(0, 0, 2), wed(9, 0).AddDate(0, 0, 5)},
	}
	for _, tt := range tests {
		if got := tt.repeat.next(tt.due, tt.now); !got.Equal(tt.want) {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestFireDue(t *testing.T) {
	var l List
	oven := l.Add(Reminder{Message: "oven", Due: wed(9, 0)})
	later := l.Add(Reminder{Message: "later", Due: wed(11, 0)})
	water := l.Add(Reminder{Message: "water", Due: wed(10, 0), Repeat: &Repeat{Every: 2 * time.Hour}})
	l.Add(Reminder{Message: "old", Due: wed(8, 0), Done: true})

	fired := l.FireDue(wed(10, 0))

	if len(fired) != 2 || fired[0].ID != oven.ID || fired[1].ID != water.ID {
		t.Fatalf("fired %+v, want oven and water", fired)
	}
	if r, _ := l.Get(oven.ID); !r.Done {
		t.Error("a one-time reminder should be done after firing")
	}
	if r, _ := l.Get(water.ID); !r.Due.Equal(wed(12, 0)) || r.Done {
		t.Errorf("repeating reminder moved to %v, want 12:00", r.Due)
	}
	if r, _ := l.Get(later.ID); r.Done || !r.FiredAt.IsZero() {
		t.Error("a future reminder should not fire")
	}
	if l.HasDue(wed(10, 0)) {
		t.Error("nothing should be due after firing")
	}
	if got := l.TakeInbox(); len(got) != 2 || len(l.Inbox) != 0 {
		t.Errorf("inbox had %d entries, want 2 then empty", len(got))
	}
}

func TestSnoozeRevivesDoneReminder(t *testing.T) {
	r := Reminder{Due: wed(9, 0), Done: true}
	r.Snooze(10*time.Minute, wed(10, 0))
	if r.Done || !r.Due.Equal(wed(10, 10)) {
		t.Errorf("got due %v done %v, want 10:10 and not done", r.Due, r.Done)
	}
}
