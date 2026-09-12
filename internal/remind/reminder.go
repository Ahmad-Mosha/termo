// Package remind holds termo's reminders: when they fire, how they repeat,
// and where they are saved.
package remind

import (
	"fmt"
	"slices"
	"time"
)

// Reminder is a message that fires at Due, once or on a schedule.
type Reminder struct {
	ID      int       `json:"id"`
	Message string    `json:"message"`
	Due     time.Time `json:"due"`
	Repeat  *Repeat   `json:"repeat,omitempty"` // nil for one-time reminders
	Done    bool      `json:"done,omitempty"`   // a one-time reminder that already fired
	FiredAt time.Time `json:"fired_at,omitzero"`
	Created time.Time `json:"created"`
}

// Repeat is a schedule: a fixed interval, or a clock time on some weekdays.
type Repeat struct {
	Every time.Duration  `json:"every,omitempty"`
	Days  []time.Weekday `json:"days,omitempty"`
	At    int            `json:"at,omitempty"` // minutes after midnight
}

// DefaultSnooze is how long a snooze lasts unless you say otherwise.
const DefaultSnooze = 10 * time.Minute

// First returns when a new reminder on this schedule should first fire.
func (s Repeat) First(now time.Time) time.Time {
	return s.next(now, now)
}

// Snooze makes r fire again d from now.
func (r *Reminder) Snooze(d time.Duration, now time.Time) {
	r.Due = now.Add(d)
	r.Done = false
}

// fire records that r went off at now and moves it to its next occurrence.
func (r *Reminder) fire(now time.Time) {
	r.FiredAt = now
	if r.Repeat == nil {
		r.Done = true
		return
	}
	r.Due = r.Repeat.next(r.Due, now)
	if r.Due.IsZero() { // a broken schedule must not fire every second
		r.Done = true
	}
}

// next returns the first time after now that the schedule fires, given that
// it last fired at due. Interval schedules keep their cadence.
func (s Repeat) next(due, now time.Time) time.Time {
	if s.Every > 0 {
		next := due.Add(s.Every)
		if !next.After(now) {
			// Skip the occurrences missed while the machine was asleep.
			next = next.Add((now.Sub(next)/s.Every + 1) * s.Every)
		}
		return next
	}
	for i := range 8 {
		t := atClock(now.AddDate(0, 0, i), s.At)
		if t.After(now) && slices.Contains(s.Days, t.Weekday()) {
			return t
		}
	}
	return time.Time{}
}

// atClock returns the given day at minutes after midnight.
func atClock(day time.Time, minutes int) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(), minutes/60, minutes%60, 0, 0, day.Location())
}

// inboxLimit caps fired reminders kept for the terminal, in case the shell
// hook is never installed to read them.
const inboxLimit = 20

// List is everything the store saves.
type List struct {
	NextID    int        `json:"next_id"`
	Reminders []Reminder `json:"reminders"`
	Inbox     []Fired    `json:"inbox,omitempty"` // fired, not yet shown in the terminal
}

// Fired is one delivery of a reminder.
type Fired struct {
	ID      int       `json:"id"`
	Message string    `json:"message"`
	Due     time.Time `json:"due"`
	At      time.Time `json:"at"`
}

// Add stores r under a new ID and returns the stored copy.
func (l *List) Add(r Reminder) Reminder {
	r.ID = max(l.NextID, 1)
	l.NextID = r.ID + 1
	l.Reminders = append(l.Reminders, r)
	return r
}

// Get returns the reminder with the given ID.
func (l *List) Get(id int) (*Reminder, error) {
	for i := range l.Reminders {
		if l.Reminders[i].ID == id {
			return &l.Reminders[i], nil
		}
	}
	return nil, fmt.Errorf("there's no reminder #%d", id)
}

// Remove deletes the reminder with the given ID and returns it.
func (l *List) Remove(id int) (Reminder, error) {
	r, err := l.Get(id)
	if err != nil {
		return Reminder{}, err
	}
	removed := *r
	l.Reminders = slices.DeleteFunc(l.Reminders, func(r Reminder) bool { return r.ID == id })
	return removed, nil
}

// ClearDone deletes the one-time reminders that already fired and reports
// how many it removed.
func (l *List) ClearDone() int {
	before := len(l.Reminders)
	l.Reminders = slices.DeleteFunc(l.Reminders, func(r Reminder) bool { return r.Done })
	return before - len(l.Reminders)
}

// HasDue reports whether any reminder should fire at now.
func (l *List) HasDue(now time.Time) bool {
	return slices.ContainsFunc(l.Reminders, func(r Reminder) bool { return !r.Done && !r.Due.After(now) })
}

// FireDue fires every reminder due at now, queues them for the terminal and
// returns them.
func (l *List) FireDue(now time.Time) []Fired {
	var fired []Fired
	for i := range l.Reminders {
		r := &l.Reminders[i]
		if r.Done || r.Due.After(now) {
			continue
		}
		fired = append(fired, Fired{ID: r.ID, Message: r.Message, Due: r.Due, At: now})
		r.fire(now)
	}
	l.Inbox = append(l.Inbox, fired...)
	if len(l.Inbox) > inboxLimit {
		l.Inbox = l.Inbox[len(l.Inbox)-inboxLimit:]
	}
	return fired
}

// TakeInbox returns the fired reminders not yet shown and empties the inbox.
func (l *List) TakeInbox() []Fired {
	in := l.Inbox
	l.Inbox = nil
	return in
}
