// Package storage decides where termo keeps its files and how it writes them
// safely when several termo processes run at once.
package storage

import (
	"os"
	"path/filepath"
	"syscall"
)

// DataDir returns termo's data directory, creating it if needed.
// It follows XDG_DATA_HOME and defaults to ~/.local/share/termo.
func DataDir() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(base, "termo")
	return dir, os.MkdirAll(dir, 0o700)
}

// Lock takes an flock(2) lock on path, creating the file if needed.
// how is syscall.LOCK_SH or syscall.LOCK_EX, optionally with LOCK_NB.
// Closing the returned file releases the lock.
func Lock(path string, how int) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), how); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

// WriteFile replaces path with data atomically, so readers never see a
// half-written file and a crash leaves the old version intact.
func WriteFile(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op once renamed
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
