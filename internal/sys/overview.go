package sys

import (
	"time"

	"github.com/shirou/gopsutil/v4/host"
)

// Overview is a one-screen snapshot of the machine's health.
type Overview struct {
	CPU    CPU           `json:"cpu"`
	Memory Memory        `json:"memory"`
	Disks  []Disk        `json:"disks"`
	Uptime time.Duration `json:"uptime"` // nanoseconds, like every other duration in termo's JSON
}

// GetOverview gathers a snapshot of CPU, memory, disk and uptime.
func GetOverview() (Overview, error) {
	c, err := CPUUsage()
	if err != nil {
		return Overview{}, err
	}
	m, err := MemoryUsage()
	if err != nil {
		return Overview{}, err
	}
	disks, err := Disks()
	if err != nil {
		return Overview{}, err
	}
	var uptime time.Duration
	if secs, err := host.Uptime(); err == nil {
		uptime = time.Duration(secs) * time.Second
	}
	return Overview{CPU: c, Memory: m, Disks: disks, Uptime: uptime}, nil
}
