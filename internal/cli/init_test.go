package cli

import (
	"os/exec"
	"strings"
	"testing"
)

// TestHooksParse syntax-checks each shell hook with shells that are installed.
func TestHooksParse(t *testing.T) {
	parseOnly := map[string]string{"zsh": "-n", "bash": "-n", "fish": "--no-execute"}
	for shell, hook := range hooks {
		path, err := exec.LookPath(shell)
		if err != nil {
			t.Logf("%s isn't installed; skipping", shell)
			continue
		}
		cmd := exec.Command(path, parseOnly[shell])
		cmd.Stdin = strings.NewReader(hook)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("%s hook doesn't parse: %v\n%s", shell, err, out)
		}
	}
}
