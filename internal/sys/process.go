package sys

import (
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// Proc is one running process's resource usage, sampled over a short
// window so CPU reflects what it's doing right now, not its average
// since it started.
type Proc struct {
	PID    int32
	Name   string
	User   string
	CPU    float64 // percent of one core, 0-100*cores
	RSS    uint64  // bytes
	MemPct float32
}

// Processes samples every running process's CPU usage over a short window
// and returns them, busiest first.
func Processes() ([]Proc, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	before := make([]float64, len(procs))
	sampled := make([]bool, len(procs))
	for i, p := range procs {
		if t, err := p.Times(); err == nil {
			before[i], sampled[i] = t.Total(), true
		}
	}
	time.Sleep(sampleWindow)

	out := make([]Proc, 0, len(procs))
	for i, p := range procs {
		if !sampled[i] {
			continue // couldn't read it before the window either; skip rather than guess
		}
		t, err := p.Times()
		if err != nil {
			continue // it exited during the window
		}
		name, err := p.Name()
		if err != nil {
			continue
		}
		user, _ := p.Username()
		memPct, _ := p.MemoryPercent()
		var rss uint64
		if mi, err := p.MemoryInfo(); err == nil && mi != nil {
			rss = mi.RSS
		}
		cpu := 100 * (t.Total() - before[i]) / sampleWindow.Seconds()
		if cpu < 0 {
			cpu = 0 // the PID was reused between samples
		}
		out = append(out, Proc{PID: p.Pid, Name: name, User: user, CPU: cpu, RSS: rss, MemPct: memPct})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CPU > out[j].CPU })
	return out, nil
}

// AppMemory is the memory used by every process belonging to one app.
type AppMemory struct {
	Name  string
	RSS   uint64 // bytes, summed across every process grouped under Name
	Procs int
}

// MemoryByApp groups every running process's memory by app, busiest
// first, so "Chrome Helper (Renderer)" and "Chrome Helper (GPU)" show up
// as one "Chrome" entry instead of forty separate rows.
func MemoryByApp() ([]AppMemory, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}
	byName := make(map[string]*AppMemory)
	for _, p := range procs {
		name, err := p.Name()
		if err != nil || name == "" {
			continue
		}
		mi, err := p.MemoryInfo()
		if err != nil || mi == nil || mi.RSS == 0 {
			// On macOS, reading another user's memory info fails silently
			// (no error, a zeroed result) rather than erroring, so a 0
			// means "couldn't read it", not "uses no memory".
			continue
		}
		app := appName(name)
		a, ok := byName[app]
		if !ok {
			a = &AppMemory{Name: app}
			byName[app] = a
		}
		a.RSS += mi.RSS
		a.Procs++
	}
	out := make([]AppMemory, 0, len(byName))
	for _, a := range byName {
		out = append(out, *a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RSS > out[j].RSS })
	return out, nil
}

// appName groups an app's helper/renderer subprocesses under its own
// name: "Google Chrome Helper (Renderer)" -> "Google Chrome",
// "Code Helper (Plugin)" -> "Code".
func appName(name string) string {
	if i := strings.Index(name, " Helper"); i != -1 {
		return strings.TrimSpace(name[:i])
	}
	return name
}
