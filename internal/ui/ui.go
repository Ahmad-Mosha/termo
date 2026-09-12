// Package ui renders termo's terminal output: colors, symbols, tables and
// friendly times. Colors are stripped when output isn't a terminal or when
// NO_COLOR is set.
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// Colors are mid-tones so they read well on light and dark terminals.
var (
	Faint  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	Accent = lipgloss.NewStyle().Foreground(lipgloss.Color("#8B5CF6"))
	Green  = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	Yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
	Red    = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	Bold   = lipgloss.NewStyle().Bold(true)
)

// Writer wraps w so colors suit whatever is behind it: a terminal, a pipe
// or a file.
func Writer(w io.Writer) io.Writer {
	return colorprofile.NewWriter(w, os.Environ())
}

// Success prints a green check and a message.
func Success(w io.Writer, format string, a ...any) {
	fmt.Fprintln(w, Green.Render("✓")+" "+fmt.Sprintf(format, a...))
}

// Warning prints a yellow exclamation mark and a message.
func Warning(w io.Writer, format string, a ...any) {
	fmt.Fprintln(w, Yellow.Render("!")+" "+fmt.Sprintf(format, a...))
}

// Table prints rows as aligned columns under a faint header. Cells may
// already be styled.
func Table(w io.Writer, header []string, rows [][]string) {
	widths := make([]int, len(header))
	for _, row := range append([][]string{header}, rows...) {
		for i, cell := range row {
			widths[i] = max(widths[i], lipgloss.Width(cell))
		}
	}
	line := func(cells []string) string {
		var b strings.Builder
		for i, cell := range cells {
			b.WriteString("  " + cell + strings.Repeat(" ", widths[i]-lipgloss.Width(cell)))
		}
		return strings.TrimRight(b.String(), " ")
	}
	fmt.Fprintln(w, Faint.Render(line(header)))
	for _, row := range rows {
		fmt.Fprintln(w, line(row))
	}
}
