package cli

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Ahmad-Mosha/termo/internal/ui"
)

func printPortsTable(w io.Writer, rows []portInfo) {
	now := time.Now()
	table := make([][]string, len(rows))
	for i, r := range rows {
		table[i] = []string{
			ui.Accent.Render(strconv.Itoa(int(r.Port))),
			orFaint(r.Process),
			pidLabel(r),
			whereLabel(r),
			upLabel(r, now),
		}
	}
	ui.Table(w, []string{"PORT", "PROCESS", "PID", "WHERE", "UP"}, table)
}

func pidLabel(r portInfo) string {
	if r.Container != "" || r.PID == 0 {
		return ui.Faint.Render("—")
	}
	return strconv.Itoa(int(r.PID))
}

func whereLabel(r portInfo) string {
	switch {
	case r.Container != "":
		return ui.Faint.Render("docker: ") + r.Container
	case r.Dir != "" && r.Branch != "":
		return ui.Path(r.Dir) + ui.Faint.Render(" ("+r.Branch+")")
	case r.Dir != "":
		return ui.Path(r.Dir)
	default:
		return ui.Faint.Render("—")
	}
}

// upLabel prefers docker's own words ("Up 2 hours") for a container, since
// termo doesn't track when the container itself started.
func upLabel(r portInfo, now time.Time) string {
	switch {
	case r.Status != "":
		return strings.TrimPrefix(r.Status, "Up ")
	case !r.Since.IsZero():
		return ui.Duration(now.Sub(r.Since))
	default:
		return ui.Faint.Render("—")
	}
}

func orFaint(s string) string {
	if s == "" {
		return ui.Faint.Render("—")
	}
	return s
}
