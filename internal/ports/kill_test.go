package ports

import (
	"os/exec"
	"testing"
	"time"
)

// TestKillProcess spawns a real process and terminates it for real, rather
// than mocking gopsutil.
func TestKillProcess(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	pid := int32(cmd.Process.Pid)

	if err := killProcess(pid, false); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		t.Fatal("the process didn't exit after SIGTERM")
	}
}

func TestKillProcessUnknownPID(t *testing.T) {
	if err := killProcess(1<<30, false); err == nil {
		t.Error("expected an error for a PID that doesn't exist")
	}
}

func TestKillNothingListening(t *testing.T) {
	// Port 1 is a privileged port nothing in this test binds to.
	if err := Kill(1, false); err == nil {
		t.Error("expected an error when nothing is listening")
	}
}
