package ui

import (
	"fmt"
	"strings"
)

// Bytes formats a byte count for people: "512 B", "3.4 MB", "1.2 GB".
func Bytes(n uint64) string {
	const unit = 1000.0
	units := [...]string{"B", "KB", "MB", "GB", "TB", "PB"}
	f, i := float64(n), 0
	for f >= unit && i < len(units)-1 {
		f /= unit
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%d B", n)
	}
	return fmt.Sprintf("%.1f %s", f, units[i])
}

// Bar renders a filled/empty block bar for percent (0-100), width cells
// wide, colored green/yellow/red by how full it is.
func Bar(percent float64, width int) string {
	percent = min(max(percent, 0), 100)
	filled := int(percent / 100 * float64(width))
	style := Green
	switch {
	case percent >= 90:
		style = Red
	case percent >= 70:
		style = Yellow
	}
	return style.Render(strings.Repeat("█", filled)) + Faint.Render(strings.Repeat("░", width-filled))
}
