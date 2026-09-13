package cli

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Ahmad-Mosha/termo/internal/sys"
	"github.com/Ahmad-Mosha/termo/internal/ui"
)

const barWidth = 20

func printOverview(w io.Writer, o sys.Overview) {
	fmt.Fprintf(w, "  %s  %5.1f%%  %s  load %s %s\n",
		ui.Bold.Render("CPU   "), o.CPU.Percent, ui.Bar(o.CPU.Percent, barWidth), loadLabel(o.CPU),
		ui.Faint.Render(fmt.Sprintf("(%d cores)", len(o.CPU.PerCore))))
	fmt.Fprintf(w, "  %s  %5.1f%%  %s  %s / %s used\n",
		ui.Bold.Render("Memory"), o.Memory.UsedPercent, ui.Bar(o.Memory.UsedPercent, barWidth), ui.Bytes(o.Memory.Used), ui.Bytes(o.Memory.Total))
	for _, d := range o.Disks {
		fmt.Fprintf(w, "  %s  %5.1f%%  %s  %s / %s  %s\n",
			ui.Bold.Render("Disk  "), d.UsedPercent, ui.Bar(d.UsedPercent, barWidth), ui.Bytes(d.Used), ui.Bytes(d.Total), ui.Faint.Render(d.Mountpoint))
	}
	fmt.Fprintf(w, "  %s  %s\n", ui.Bold.Render("Uptime"), ui.Duration(o.Uptime))
}

func loadLabel(c sys.CPU) string {
	return fmt.Sprintf("%.2f %.2f %.2f", c.Load1, c.Load5, c.Load15)
}

func printCPU(w io.Writer, c sys.CPU) {
	fmt.Fprintf(w, "  Overall  %5.1f%%  %s\n", c.Percent, ui.Bar(c.Percent, barWidth))
	fmt.Fprintf(w, "  Load     %s  %s\n", loadLabel(c), ui.Faint.Render(fmt.Sprintf("(1m 5m 15m, %d cores)", len(c.PerCore))))
	if len(c.PerCore) == 0 {
		return
	}
	fmt.Fprintln(w)
	rows := make([][]string, len(c.PerCore))
	for i, p := range c.PerCore {
		rows[i] = []string{strconv.Itoa(i), fmt.Sprintf("%5.1f%%  %s", p, ui.Bar(p, barWidth))}
	}
	ui.Table(w, []string{"CORE", "USAGE"}, rows)
}

func printMemory(w io.Writer, m sys.Memory, apps []sys.AppMemory) {
	fmt.Fprintf(w, "  Used       %5.1f%%  %s  %s / %s\n", m.UsedPercent, ui.Bar(m.UsedPercent, barWidth), ui.Bytes(m.Used), ui.Bytes(m.Total))
	fmt.Fprintf(w, "  Available  %s\n", ui.Bytes(m.Available))
	if m.SwapTotal > 0 {
		fmt.Fprintf(w, "  Swap       %s / %s\n", ui.Bytes(m.SwapUsed), ui.Bytes(m.SwapTotal))
	}
	if len(apps) == 0 {
		return
	}
	fmt.Fprintln(w)
	rows := make([][]string, len(apps))
	for i, a := range apps {
		rows[i] = []string{a.Name, ui.Bytes(a.RSS), ui.Faint.Render(strconv.Itoa(a.Procs))}
	}
	ui.Table(w, []string{"APP", "MEMORY", "PROCS"}, rows)
}

func printDisks(w io.Writer, disks []sys.Disk) {
	if len(disks) == 0 {
		fmt.Fprintln(w, "No disks found.")
		return
	}
	rows := make([][]string, len(disks))
	for i, d := range disks {
		rows[i] = []string{d.Mountpoint, ui.Bytes(d.Used), ui.Bytes(d.Total), fmt.Sprintf("%s %.1f%%", ui.Bar(d.UsedPercent, barWidth), d.UsedPercent)}
	}
	ui.Table(w, []string{"MOUNT", "USED", "TOTAL", "USAGE"}, rows)
}

func printNetwork(w io.Writer, ifaces []sys.NetIO) {
	if len(ifaces) == 0 {
		fmt.Fprintln(w, "No active network interfaces.")
		return
	}
	rows := make([][]string, len(ifaces))
	for i, n := range ifaces {
		rows[i] = []string{
			n.Interface,
			ui.Bytes(n.SentPerSec) + "/s",
			ui.Bytes(n.RecvPerSec) + "/s",
			ui.Faint.Render(ui.Bytes(n.TotalSent)),
			ui.Faint.Render(ui.Bytes(n.TotalRecv)),
		}
	}
	ui.Table(w, []string{"INTERFACE", "SENT/S", "RECV/S", "TOTAL SENT", "TOTAL RECV"}, rows)
}

func printProcesses(w io.Writer, procs []sys.Proc) {
	if len(procs) == 0 {
		fmt.Fprintln(w, "No processes found.")
		return
	}
	rows := make([][]string, len(procs))
	for i, p := range procs {
		mem := ui.Faint.Render("—")
		if p.RSS > 0 {
			mem = ui.Bytes(p.RSS)
		}
		rows[i] = []string{
			strconv.Itoa(int(p.PID)),
			p.Name,
			fmt.Sprintf("%.1f%%", p.CPU),
			mem,
			ui.Faint.Render(p.User),
		}
	}
	ui.Table(w, []string{"PID", "NAME", "CPU", "MEM", "USER"}, rows)
}
