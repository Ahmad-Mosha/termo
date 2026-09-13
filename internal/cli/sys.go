package cli

import (
	"encoding/json"
	"sort"

	"github.com/Ahmad-Mosha/termo/internal/sys"
	"github.com/Ahmad-Mosha/termo/internal/ui"
	"github.com/spf13/cobra"
)

func sysCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "sys",
		Short: "System health at a glance",
		Long: `Show CPU, memory, disk, network and process usage.
Run it with no subcommand for a one-screen overview.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			o, err := sys.GetOverview()
			if err != nil {
				return err
			}
			if asJSON {
				return encodeJSON(cmd, o)
			}
			printOverview(ui.Writer(cmd.OutOrStdout()), o)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON for scripts")
	cmd.AddCommand(sysCPUCmd(), sysMemCmd(), sysDiskCmd(), sysNetCmd(), sysPSCmd())
	return cmd
}

func sysCPUCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "cpu",
		Short: "Overall and per-core CPU usage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := sys.CPUUsage()
			if err != nil {
				return err
			}
			if asJSON {
				return encodeJSON(cmd, c)
			}
			printCPU(ui.Writer(cmd.OutOrStdout()), c)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON for scripts")
	return cmd
}

func sysMemCmd() *cobra.Command {
	var asJSON bool
	var limit int
	cmd := &cobra.Command{
		Use:     "mem",
		Aliases: []string{"memory"},
		Short:   "Memory usage, grouped by app",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			m, err := sys.MemoryUsage()
			if err != nil {
				return err
			}
			apps, err := sys.MemoryByApp()
			if err != nil {
				return err
			}
			apps = firstN(apps, limit)
			if asJSON {
				return encodeJSON(cmd, struct {
					Memory sys.Memory      `json:"memory"`
					Apps   []sys.AppMemory `json:"apps"`
				}{m, apps})
			}
			printMemory(ui.Writer(cmd.OutOrStdout()), m, apps)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON for scripts")
	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "how many apps to show")
	return cmd
}

func sysDiskCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "disk",
		Aliases: []string{"df"},
		Short:   "Disk usage",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			disks, err := sys.Disks()
			if err != nil {
				return err
			}
			if asJSON {
				return encodeJSON(cmd, disks)
			}
			printDisks(ui.Writer(cmd.OutOrStdout()), disks)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON for scripts")
	return cmd
}

func sysNetCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "net",
		Short: "Network throughput",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ifaces, err := sys.Network()
			if err != nil {
				return err
			}
			if asJSON {
				return encodeJSON(cmd, ifaces)
			}
			printNetwork(ui.Writer(cmd.OutOrStdout()), ifaces)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON for scripts")
	return cmd
}

func sysPSCmd() *cobra.Command {
	var asJSON, byMem bool
	var limit int
	cmd := &cobra.Command{
		Use:     "ps",
		Aliases: []string{"top"},
		Short:   "Running processes",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			procs, err := sys.Processes()
			if err != nil {
				return err
			}
			if byMem {
				sort.Slice(procs, func(i, j int) bool { return procs[i].MemPct > procs[j].MemPct })
			}
			procs = firstN(procs, limit)
			if asJSON {
				return encodeJSON(cmd, procs)
			}
			printProcesses(ui.Writer(cmd.OutOrStdout()), procs)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON for scripts")
	cmd.Flags().BoolVarP(&byMem, "mem", "m", false, "sort by memory instead of CPU")
	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "how many processes to show")
	return cmd
}

// firstN returns the first n items of s, or all of it if there are fewer.
func firstN[T any](s []T, n int) []T {
	if n < len(s) {
		return s[:n]
	}
	return s
}

// encodeJSON prints v as indented JSON.
func encodeJSON(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
