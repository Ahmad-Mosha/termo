package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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
	cmd.AddCommand(portsKillCmd())
	return cmd
}

func portsKillCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:     "kill <port>",
		Aliases: []string{"stop"},
		Short:   "Stop whatever is listening on a port",
		Long: `Stop whatever is listening on a port: the process, or the docker
container if one owns it.

By default this asks nicely (SIGTERM, or docker stop). --force stops it
right away (SIGKILL, or docker kill).`,
		Example: `termo ports kill 3000
termo ports kill 3000 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			port, err := parsePort(args[0])
			if err != nil {
				return err
			}
			row, err := findPort(port)
			if err != nil {
				return err
			}
			if err := ports.Kill(port, force); err != nil {
				return err
			}
			verb := "Stopped"
			if force {
				verb = "Killed"
			}
			ui.Success(ui.Writer(cmd.OutOrStdout()), "%s port %s %s",
				verb, ui.Accent.Render(strconv.Itoa(int(port))), ui.Faint.Render("("+describe(row)+")"))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "stop it right away instead of asking nicely")
	return cmd
}

func parsePort(s string) (uint32, error) {
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil || n == 0 || n > 65535 {
		return 0, fmt.Errorf("%q isn't a port number", s)
	}
	return uint32(n), nil
}

// findPort looks up one port among what's currently listening, so a
// command can report what it's about to act on.
func findPort(port uint32) (portInfo, error) {
	rows, err := gatherPorts()
	if err != nil {
		return portInfo{}, err
	}
	for _, r := range rows {
		if r.Port == port {
			return r, nil
		}
	}
	return portInfo{}, errors.New("nothing is listening on port " + strconv.Itoa(int(port)))
}

// describe names what's behind a port, for a confirmation message.
func describe(r portInfo) string {
	switch {
	case r.Container != "":
		return "docker: " + r.Container
	case r.Process != "":
		return r.Process
	default:
		return fmt.Sprintf("pid %d", r.PID)
	}
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
