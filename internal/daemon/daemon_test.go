package daemon

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	"github.com/Ahmad-Mosha/termo/internal/remind"
)

func TestRunFiresDueReminders(t *testing.T) {
	dir := t.TempDir()
	store := remind.NewStore(dir)
	store.Update(func(l *remind.List) error {
		l.Add(remind.Reminder{Message: "stretch", Due: time.Now().Add(-time.Minute)})
		l.Add(remind.Reminder{Message: "later", Due: time.Now().Add(time.Hour)})
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	delivered := make(chan remind.Fired, 2)
	done := make(chan error)
	go func() {
		done <- Run(ctx, dir, func(f remind.Fired) error { delivered <- f; return nil }, log.New(io.Discard, "", 0))
	}()

	select {
	case f := <-delivered:
		if f.Message != "stretch" {
			t.Errorf("delivered %q, want stretch", f.Message)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("nothing was delivered")
	}
	if _, ok := Running(dir); !ok {
		t.Error("Running should report the daemon")
	}
	if err := Run(ctx, dir, nil, log.New(io.Discard, "", 0)); err != ErrRunning {
		t.Errorf("a second daemon got %v, want ErrRunning", err)
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if _, ok := Running(dir); ok {
		t.Error("Running should be false after the daemon stops")
	}
	store.View(func(l *remind.List) error {
		if r, _ := l.Get(1); !r.Done {
			t.Error("the fired reminder should be done")
		}
		if len(l.Inbox) != 1 {
			t.Errorf("inbox has %d entries, want 1", len(l.Inbox))
		}
		return nil
	})
	if len(delivered) != 0 {
		t.Error("the future reminder should not fire")
	}
}
