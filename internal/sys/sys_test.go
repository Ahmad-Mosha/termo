package sys

import "testing"

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
