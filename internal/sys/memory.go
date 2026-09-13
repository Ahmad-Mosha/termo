package sys

import "github.com/shirou/gopsutil/v4/mem"

// Memory is overall RAM and swap usage.
type Memory struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Available   uint64  `json:"available"`
	UsedPercent float64 `json:"used_percent"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
}

// MemoryUsage reads current memory and swap usage.
func MemoryUsage() (Memory, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return Memory{}, err
	}
	m := Memory{Total: vm.Total, Used: vm.Used, Available: vm.Available, UsedPercent: vm.UsedPercent}
	if sw, err := mem.SwapMemory(); err == nil {
		m.SwapTotal, m.SwapUsed = sw.Total, sw.Used
	}
	return m, nil
}
