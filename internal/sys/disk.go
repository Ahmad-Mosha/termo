package sys

import (
	"sort"
	"strings"

	"github.com/shirou/gopsutil/v4/disk"
)

// Disk is how full one mounted filesystem is.
type Disk struct {
	Mountpoint  string  `json:"mountpoint"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

// Disks reports usage for the real filesystems on this machine, busiest
// first. macOS mounts several synthetic system volumes (Preboot, VM,
// Update, Data, ...) alongside the real root, all sharing one physical
// APFS container and all reporting that whole container's free space; so
// "/" already gives the container's true usage and the rest are skipped
// as noisy duplicates, not because they're uninteresting on their own.
func Disks() ([]Disk, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}
	var out []Disk
	for _, p := range parts {
		if skipMount(p) {
			continue
		}
		u, err := disk.Usage(p.Mountpoint)
		if err != nil || u.Total == 0 {
			continue
		}
		out = append(out, Disk{Mountpoint: p.Mountpoint, Total: u.Total, Used: u.Used, UsedPercent: u.UsedPercent})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Used > out[j].Used })
	return out, nil
}

func skipMount(p disk.PartitionStat) bool {
	switch p.Fstype {
	case "devfs", "autofs":
		return true
	}
	// macOS mounts several synthetic system volumes under here (Preboot,
	// VM, Update, ...); "/" already reports the same container's usage.
	return strings.HasPrefix(p.Mountpoint, "/System/Volumes/")
}
