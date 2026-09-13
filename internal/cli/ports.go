package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Ahmad-Mosha/termo/internal/ports"
	"github.com/Ahmad-Mosha/termo/internal/ui"
	"github.com/spf13/cobra"
)

func portsCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "ports",
		Aliases: []string{"p"},
		Short:   "See what's listening on your ports",
		Long: `List every port your machine is listening on: which process owns it,
which project or docker container it belongs to, and how long it's been up.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return listPorts(cmd, asJSON) },
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON for scripts")
	return cmd
}

// portInfo is one listening port, and whatever termo could learn about
// what's behind it.
type portInfo struct {
	Port      uint32    `json:"port"`
	PID       int32     `json:"pid,omitempty"`
	Process   string    `json:"process,omitempty"`
	Dir       string    `json:"dir,omitempty"`
	Branch    string    `json:"branch,omitempty"`
	Container string    `json:"container,omitempty"`
	Since     time.Time `json:"since,omitzero"`
	Status    string    `json:"status,omitempty"` // docker's own words, for a container
}

// gatherPorts lists listening ports and fills in each one's project or
// docker container.
func gatherPorts() ([]portInfo, error) {
	list, err := ports.List()
	if err != nil {
		return nil, err
	}
	byPort, err := ports.Containers()
	if err != nil {
		return nil, err
	}

	rows := make([]portInfo, len(list))
	for i, p := range list {
		info := portInfo{Port: p.Number, PID: p.PID, Process: p.Process, Since: p.Since}
		switch c, ok := byPort[p.Number]; {
		case ok:
			// Docker owns the socket (com.docker.backend on macOS, say),
			// so the container is what matters, not the host process.
			info.PID, info.Process = 0, imageRepo(c.Image)
			info.Container, info.Status, info.Since = c.Name, c.Status, time.Time{}
		case p.Cwd != "":
			proj := ports.ProjectFor(p.Cwd)
			info.Dir, info.Branch = proj.Dir, proj.Branch
		}
		rows[i] = info
	}
	return rows, nil
}

func listPorts(cmd *cobra.Command, asJSON bool) error {
	rows, err := gatherPorts()
	if err != nil {
		return err
	}
	if asJSON {
		if rows == nil {
			rows = []portInfo{}
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}

	w := ui.Writer(cmd.OutOrStdout())
	if len(rows) == 0 {
		fmt.Fprintln(w, "Nothing is listening.")
		return nil
	}
	printPortsTable(w, rows)
	return nil
}

// imageRepo turns an image reference into a short, friendly name:
// "redis:alpine" -> "redis", "ghcr.io/acme/api:latest" -> "api".
func imageRepo(image string) string {
	if i := strings.LastIndexByte(image, '/'); i != -1 {
		image = image[i+1:]
	}
	if i := strings.IndexAny(image, ":@"); i != -1 {
		image = image[:i]
	}
	return image
}
