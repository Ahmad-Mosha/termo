// Package ports finds what's listening on your machine's ports, where it
// lives, and lets you stop it.
package ports

import (
	"sort"
	"time"

	psnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// Port is one listening TCP port.
type Port struct {
	Number  uint32
	PID     int32     // 0 if no process could be matched
	Process string    // empty if the process is gone or unreadable
	Cwd     string    // empty if unknown or unreadable
	Since   time.Time // zero if unknown
}

// List returns every TCP port your machine is listening on, one row per
// port even if a process binds it on both IPv4 and IPv6.
func List() ([]Port, error) {
	conns, err := psnet.Connections("tcp")
	if err != nil {
		return nil, err
	}

	seen := make(map[uint32]bool)
	procs := make(map[int32]*process.Process) // avoid looking up the same PID twice
	var out []Port

	for _, c := range conns {
		if c.Status != "LISTEN" || seen[c.Laddr.Port] {
			continue
		}
		seen[c.Laddr.Port] = true
		p := Port{Number: c.Laddr.Port, PID: c.Pid}

		if c.Pid > 0 {
			proc, ok := procs[c.Pid]
			if !ok {
				proc, _ = process.NewProcess(c.Pid)
				procs[c.Pid] = proc
			}
			if proc != nil {
				if name, err := proc.Name(); err == nil {
					p.Process = name
				}
				if cwd, err := proc.Cwd(); err == nil {
					p.Cwd = cwd
				}
				if ms, err := proc.CreateTime(); err == nil {
					p.Since = time.UnixMilli(ms)
				}
			}
		}
		out = append(out, p)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}
