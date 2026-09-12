package remind

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/Ahmad-Mosha/termo/internal/storage"
)

// Store saves the reminder list as JSON. Every read and write holds a file
// lock, so the CLI and the daemon can use it at the same time.
type Store struct {
	path string
}

// NewStore returns the store kept in dir.
func NewStore(dir string) *Store {
	return &Store{path: filepath.Join(dir, "reminders.json")}
}

// Path returns the file the reminders are saved in.
func (s *Store) Path() string { return s.path }

// View loads the list and passes it to fn without saving.
func (s *Store) View(fn func(*List) error) error {
	return s.with(syscall.LOCK_SH, fn)
}

// Update loads the list, lets fn change it and saves the result.
// Nothing is saved if fn returns an error.
func (s *Store) Update(fn func(*List) error) error {
	return s.with(syscall.LOCK_EX, func(l *List) error {
		if err := fn(l); err != nil {
			return err
		}
		data, err := json.MarshalIndent(l, "", "  ")
		if err != nil {
			return err
		}
		return storage.WriteFile(s.path, data)
	})
}

func (s *Store) with(how int, fn func(*List) error) error {
	lock, err := storage.Lock(s.path+".lock", how)
	if err != nil {
		return err
	}
	defer lock.Close()

	l := &List{NextID: 1}
	data, err := os.ReadFile(s.path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return err
	default:
		// Refuse to go on with a damaged file rather than overwrite it.
		if err := json.Unmarshal(data, l); err != nil {
			return fmt.Errorf("can't read %s: %w", s.path, err)
		}
	}
	return fn(l)
}
