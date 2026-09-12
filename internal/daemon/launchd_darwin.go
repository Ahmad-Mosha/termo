package daemon

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

const label = "com.termo.daemon"

var plist = template.Must(template.New("plist").Funcs(template.FuncMap{"xml": xmlEscape}).Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>{{xml .Label}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{xml .Exe}}</string>
		<string>daemon</string>
		<string>run</string>
	</array>
	<key>EnvironmentVariables</key>
	<dict>
		<key>XDG_DATA_HOME</key>
		<string>{{xml .DataHome}}</string>
	</dict>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>{{xml .Log}}</string>
	<key>StandardErrorPath</key>
	<string>{{xml .Log}}</string>
</dict>
</plist>
`))

// Install registers the daemon as a launchd agent that starts when you log
// in and restarts if it stops, then starts it. Running it again restarts
// the daemon, which is how to pick up a new termo binary.
func Install(dir string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if strings.Contains(exe, "go-build") {
		return errors.New("this termo binary is temporary (go run); install it first with go install")
	}
	path, err := plistPath()
	if err != nil {
		return err
	}
	var b strings.Builder
	err = plist.Execute(&b, map[string]string{
		"Label":    label,
		"Exe":      exe,
		"DataHome": filepath.Dir(dir), // so the daemon finds the same data as this shell
		"Log":      LogPath(dir),
	})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return err
	}

	stop()
	// launchd can take a moment to let go of a service it just stopped.
	var out []byte
	for range 10 {
		if out, err = exec.Command("launchctl", "bootstrap", domain(), path).CombinedOutput(); err == nil {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("launchctl bootstrap: %s", strings.TrimSpace(string(out)))
}

// Uninstall stops the daemon and removes it from launchd.
func Uninstall() error {
	path, err := plistPath()
	if err != nil {
		return err
	}
	stop()
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// Installed reports whether the launchd agent is set up.
func Installed() bool {
	path, err := plistPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// stop unloads the agent if it's loaded.
func stop() {
	exec.Command("launchctl", "bootout", domain()+"/"+label).Run()
}

func domain() string { return fmt.Sprintf("gui/%d", os.Getuid()) }

func plistPath() (string, error) {
	home, err := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", label+".plist"), err
}

func xmlEscape(s string) string {
	var b strings.Builder
	xml.EscapeText(&b, []byte(s))
	return b.String()
}
