package sys

import (
	"os/exec"
	"testing"
)

// These test real invariants on whatever machine runs them, rather than
// mocking gopsutil: the actual numbers vary machine to machine, but a
// working reading has to look a certain way.

func TestCPUUsage(t *testing.T) {
	c, err := CPUUsage()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.PerCore) == 0 {
		t.Error("PerCore is empty; every machine has at least one core")
	}
	if c.Percent < 0 || c.Percent > 100 {
		t.Errorf("Percent = %v, want 0-100", c.Percent)
	}
	for _, p := range c.PerCore {
		if p < 0 || p > 100 {
			t.Errorf("a core reported %v%%, want 0-100", p)
		}
	}
}

func TestMemoryUsage(t *testing.T) {
	m, err := MemoryUsage()
	if err != nil {
		t.Fatal(err)
	}
	if m.Total == 0 {
		t.Error("Total is 0; every machine has some RAM")
	}
	if m.Used > m.Total {
		t.Errorf("Used (%d) > Total (%d)", m.Used, m.Total)
	}
	if m.UsedPercent < 0 || m.UsedPercent > 100 {
		t.Errorf("UsedPercent = %v, want 0-100", m.UsedPercent)
	}
}

func TestDisks(t *testing.T) {
	disks, err := Disks()
	if err != nil {
		t.Fatal(err)
	}
	if len(disks) == 0 {
		t.Fatal("no disks found; every machine has at least a root filesystem")
	}
	var foundRoot bool
	for _, d := range disks {
		if d.Mountpoint == "/" {
			foundRoot = true
		}
		if d.Used > d.Total {
			t.Errorf("%s: Used (%d) > Total (%d)", d.Mountpoint, d.Used, d.Total)
		}
	}
	if !foundRoot {
		t.Errorf("no root filesystem among %+v", disks)
	}
	seen := make(map[uint64]bool)
	for _, d := range disks {
		if seen[d.Total] {
			t.Errorf("two disks report the exact same size (%d bytes); a synthetic mount slipped through", d.Total)
		}
		seen[d.Total] = true
	}
}

func TestNetwork(t *testing.T) {
	ifaces, err := Network()
	if err != nil {
		t.Fatal(err)
	}
	for i, n := range ifaces {
		if n.Interface == "lo0" || n.Interface == "lo" {
			t.Errorf("loopback %q should have been filtered out", n.Interface)
		}
		if n.TotalSent == 0 && n.TotalRecv == 0 {
			t.Errorf("%s: an interface with no traffic should have been filtered out", n.Interface)
		}
		if i > 0 {
			prev := ifaces[i-1]
			if prev.TotalSent+prev.TotalRecv < n.TotalSent+n.TotalRecv {
				t.Errorf("not sorted busiest-first: %s came before %s", prev.Interface, n.Interface)
			}
		}
	}
}

// TestProcessesFindsABusyProcess spawns a real CPU-bound process (rather
// than mocking gopsutil) and checks that it shows up under real load.
func TestProcessesFindsABusyProcess(t *testing.T) {
	cmd := exec.Command("yes") // a tight loop; Stdout left nil discards it
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	procs, err := Processes()
	if err != nil {
		t.Fatal(err)
	}
	var found *Proc
	for i := range procs {
		if procs[i].PID == int32(cmd.Process.Pid) {
			found = &procs[i]
		}
	}
	if found == nil {
		t.Fatalf("didn't find our own busy process (pid %d) among %d", cmd.Process.Pid, len(procs))
	}
	if found.CPU < 30 {
		t.Errorf("CPU = %.1f%%, want something substantial for a process in a tight loop", found.CPU)
	}
	if found.Name == "" {
		t.Error("Name is empty")
	}
}

func TestProcessesSortedBusiestFirst(t *testing.T) {
	procs, err := Processes()
	if err != nil {
		t.Fatal(err)
	}
	if len(procs) == 0 {
		t.Fatal("no processes found; at least this test process should show up")
	}
	for i := 1; i < len(procs); i++ {
		if procs[i-1].CPU < procs[i].CPU {
			t.Errorf("not sorted busiest-first at index %d: %.1f then %.1f", i, procs[i-1].CPU, procs[i].CPU)
		}
	}
}

func TestMemoryByApp(t *testing.T) {
	apps, err := MemoryByApp()
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) == 0 {
		t.Fatal("no apps found")
	}
	for i, a := range apps {
		if a.Procs < 1 {
			t.Errorf("%s: Procs = %d, want at least 1", a.Name, a.Procs)
		}
		if a.RSS == 0 {
			t.Errorf("%s: RSS is 0", a.Name)
		}
		if i > 0 && apps[i-1].RSS < a.RSS {
			t.Errorf("not sorted busiest-first: %s came before %s", apps[i-1].Name, a.Name)
		}
	}
}

func TestAppName(t *testing.T) {
	tests := map[string]string{
		"Google Chrome Helper (Renderer)": "Google Chrome",
		"Google Chrome Helper (GPU)":      "Google Chrome",
		"Code Helper (Plugin)":            "Code",
		"Discord Helper (Renderer)":       "Discord",
		"node":                            "node",
		"Finder":                          "Finder",
	}
	for in, want := range tests {
		if got := appName(in); got != want {
			t.Errorf("appName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGetOverview(t *testing.T) {
	o, err := GetOverview()
	if err != nil {
		t.Fatal(err)
	}
	if o.Uptime <= 0 {
		t.Error("Uptime should be positive; the machine has been up for a while to run this test")
	}
	if o.Memory.Total == 0 || len(o.Disks) == 0 {
		t.Errorf("incomplete overview: %+v", o)
	}
}
