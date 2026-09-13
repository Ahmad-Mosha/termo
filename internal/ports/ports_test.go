package ports

import (
	"net"
	"os"
	"testing"
	"time"
)

// TestListFindsOwnListener opens a real TCP listener and checks that List
// reports it under this process's PID, name and working directory.
func TestListFindsOwnListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := uint32(ln.Addr().(*net.TCPAddr).Port)

	got, err := List()
	if err != nil {
		t.Fatal(err)
	}
	var found *Port
	for i := range got {
		if got[i].Number == port {
			found = &got[i]
		}
	}
	if found == nil {
		t.Fatalf("List() didn't include port %d among %d ports", port, len(got))
	}
	if found.PID != int32(os.Getpid()) {
		t.Errorf("PID = %d, want this test process (%d)", found.PID, os.Getpid())
	}
	if found.Process == "" {
		t.Error("Process name is empty")
	}
	wantCwd, _ := os.Getwd()
	if found.Cwd != wantCwd {
		t.Errorf("Cwd = %q, want %q", found.Cwd, wantCwd)
	}
	if found.Since.IsZero() || time.Since(found.Since) > time.Minute {
		t.Errorf("Since = %v, want roughly now", found.Since)
	}
}

func TestListHasNoDuplicatePorts(t *testing.T) {
	got, err := List()
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[uint32]bool)
	for _, p := range got {
		if seen[p.Number] {
			t.Errorf("port %d listed twice", p.Number)
		}
		seen[p.Number] = true
	}
}
