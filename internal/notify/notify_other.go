//go:build !darwin

package notify

import "os/exec"

// Hint tells people what to check when notifications don't appear.
const Hint = "Didn't see it? Check that notify-send is installed and a notification daemon is running."

// Banner shows a desktop notification with notify-send.
func Banner(title, message string) error {
	return exec.Command("notify-send", "--app-name=termo", title, message).Run()
}
