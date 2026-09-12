// Package daemon runs termo in the background: it fires reminders when
// they're due, and installs itself so it starts when you log in.
package daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Ahmad-Mosha/termo/internal/remind"
	"github.com/Ahmad-Mosha/termo/internal/storage"
)

// ErrRunning means another daemon already holds the lock.
var ErrRunning = errors.New("the daemon is already running")

// Run fires due reminders until ctx is done, handing each one to deliver.
//
// It checks the wall clock every second instead of sleeping until the next
// reminder, because timers pause while a Mac sleeps. Anything missed during
// sleep fires once on wake.
func Run(ctx context.Context, dir string, deliver func(remind.Fired) error, logger *log.Logger) error {
	lock, err := storage.Lock(lockPath(dir), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return ErrRunning
	}
	if err != nil {
		return err
	}
	defer lock.Close()
	lock.Truncate(0)
	fmt.Fprintf(lock, "%d\n", os.Getpid())

	store := remind.NewStore(dir)
	logger.Printf("started, watching %s", store.Path())
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		fired, err := fireDue(store, time.Now())
		if err != nil {
			logger.Printf("error: %v", err)
		}
		for _, f := range fired {
			logger.Printf("⏰ %s (#%d)", f.Message, f.ID)
			if err := deliver(f); err != nil {
				logger.Printf("error: %v", err)
			}
		}
		select {
		case <-ctx.Done():
			logger.Print("stopped")
			return nil
		case <-tick.C:
		}
	}
}

// Running reports whether a daemon is running and its process ID.
func Running(dir string) (pid int, ok bool) {
	lock, err := storage.Lock(lockPath(dir), syscall.LOCK_SH|syscall.LOCK_NB)
	if err == nil {
		lock.Close() // nobody holds it
		return 0, false
	}
	data, _ := os.ReadFile(lockPath(dir))
	pid, _ = strconv.Atoi(strings.TrimSpace(string(data)))
	return pid, errors.Is(err, syscall.EWOULDBLOCK)
}

// LogPath is where the installed daemon writes its log.
func LogPath(dir string) string { return filepath.Join(dir, "daemon.log") }

func lockPath(dir string) string { return filepath.Join(dir, "daemon.lock") }

// fireDue fires whatever is due. It only takes the write lock when
// something is due, so the idle loop stays read-only.
func fireDue(store *remind.Store, now time.Time) ([]remind.Fired, error) {
	var due bool
	if err := store.View(func(l *remind.List) error { due = l.HasDue(now); return nil }); err != nil || !due {
		return nil, err
	}
	var fired []remind.Fired
	err := store.Update(func(l *remind.List) error { fired = l.FireDue(now); return nil })
	return fired, err
}
