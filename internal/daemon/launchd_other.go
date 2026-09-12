//go:build !darwin

package daemon

import "errors"

var errUnsupported = errors.New("installing the daemon only works on macOS for now; run `termo daemon run` from your init system (for example a systemd user service)")

// Install is only supported on macOS.
func Install(string) error { return errUnsupported }

// Uninstall is only supported on macOS.
func Uninstall() error { return errUnsupported }

// Installed is always false outside macOS.
func Installed() bool { return false }
