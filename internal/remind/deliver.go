package remind

import "github.com/Ahmad-Mosha/termo/internal/notify"

// Notify shows a fired reminder as a desktop notification.
func Notify(f Fired) error {
	return notify.Banner("⏰ Reminder", f.Message)
}
