// Package sys reads your machine's vital signs: CPU, memory, disk, network
// and processes.
package sys

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
)

// sampleWindow is how long a CPU or network reading blocks to see how busy
// things are right now, rather than some average since boot.
const sampleWindow = 300 * time.Millisecond

// CPU is how busy the processor is right now.
type CPU struct {
	Percent float64   `json:"percent"`  // 0-100, averaged across every core
	PerCore []float64 `json:"per_core"` // one entry per logical core
	Load1   float64   `json:"load1"`    // load average over the last minute
	Load5   float64   `json:"load5"`
	Load15  float64   `json:"load15"`
}

// CPUUsage samples CPU usage over a short window.
func CPUUsage() (CPU, error) {
	perCore, err := cpu.Percent(sampleWindow, true)
	if err != nil {
		return CPU{}, err
	}
	c := CPU{PerCore: perCore}
	for _, p := range perCore {
		c.Percent += p
	}
	if len(perCore) > 0 {
		c.Percent /= float64(len(perCore))
	}
	if avg, err := load.Avg(); err == nil {
		c.Load1, c.Load5, c.Load15 = avg.Load1, avg.Load5, avg.Load15
	}
	return c, nil
}
