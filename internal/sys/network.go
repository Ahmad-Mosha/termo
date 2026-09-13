package sys

import (
	"sort"
	"time"

	"github.com/shirou/gopsutil/v4/net"
)

// netSampleWindow is longer than the CPU one: network throughput is
// bursty, so a slightly wider window reads less noisy.
const netSampleWindow = 500 * time.Millisecond

// NetIO is one network interface's throughput, sampled just now, plus its
// lifetime totals since boot.
type NetIO struct {
	Interface  string `json:"interface"`
	SentPerSec uint64 `json:"sent_per_sec"` // bytes/sec, just now
	RecvPerSec uint64 `json:"recv_per_sec"`
	TotalSent  uint64 `json:"total_sent"` // bytes, since boot
	TotalRecv  uint64 `json:"total_recv"`
}

// Network samples every active network interface's current throughput,
// busiest first. The loopback interface and anything that's never sent or
// received a byte are left out.
func Network() ([]NetIO, error) {
	before, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}
	time.Sleep(netSampleWindow)
	after, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	byName := make(map[string]net.IOCountersStat, len(before))
	for _, b := range before {
		byName[b.Name] = b
	}
	secs := netSampleWindow.Seconds()
	var out []NetIO
	for _, a := range after {
		if a.Name == "lo0" || a.Name == "lo" || (a.BytesSent == 0 && a.BytesRecv == 0) {
			continue
		}
		b := byName[a.Name] // zero value if the interface just appeared
		out = append(out, NetIO{
			Interface:  a.Name,
			SentPerSec: perSecond(a.BytesSent, b.BytesSent, secs),
			RecvPerSec: perSecond(a.BytesRecv, b.BytesRecv, secs),
			TotalSent:  a.BytesSent,
			TotalRecv:  a.BytesRecv,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TotalSent+out[i].TotalRecv > out[j].TotalSent+out[j].TotalRecv })
	return out, nil
}

// perSecond turns a byte count into a rate, guarding against the rare
// counter reset that would otherwise underflow to a huge number.
func perSecond(after, before uint64, secs float64) uint64 {
	if after < before {
		return 0
	}
	return uint64(float64(after-before) / secs)
}
