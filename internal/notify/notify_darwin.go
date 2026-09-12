// Package notify shows desktop notifications.
package notify

import (
	"fmt"
	"os/exec"
	"strings"
)

// Hint tells people what to check when notifications don't appear.
const Hint = "Didn't see it? Allow notifications for Script Editor in System Settings → Notifications."

// Banner shows a notification in the top-right corner of the screen.
//
// macOS lists these under Script Editor in System Settings → Notifications,
// and silently drops them if Script Editor isn't allowed to notify.
func Banner(title, message string) error {
	// The text travels as arguments rather than being pasted into the
	// script, so a message can't inject AppleScript.
	cmd := exec.Command("/usr/bin/osascript", "-", title, message)
	cmd.Stdin = strings.NewReader(`on run argv
	display notification (item 2 of argv) with title (item 1 of argv) sound name "Glass"
end run`)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("osascript: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
