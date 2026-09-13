package sys

import "github.com/shirou/gopsutil/v4/mem"

// Memory is overall RAM and swap usage.
type Memory struct {
	Total, Used, Available uint64
	UsedPercent            float64
	SwapTotal, SwapUsed    uint64
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
