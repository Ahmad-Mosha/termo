package cli

import (
	"fmt"
	"log"
	"time"

	"github.com/Ahmad-Mosha/termo/internal/daemon"
	"github.com/Ahmad-Mosha/termo/internal/remind"
	"github.com/Ahmad-Mosha/termo/internal/storage"
	"github.com/Ahmad-Mosha/termo/internal/ui"
	"github.com/spf13/cobra"
)

func daemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Run termo in the background so reminders fire on time",
		Long: `The daemon checks your reminders every second and sends a notification
when one is due. Install it once and it starts whenever you log in.`,
	}
	cmd.AddCommand(daemonInstallCmd(), daemonStatusCmd(), daemonUninstallCmd(), daemonRunCmd())
	return cmd
}

func daemonInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Start the daemon now and whenever you log in",
		Long:  "Start the daemon now and whenever you log in. Run it again after updating termo to restart the daemon on the new version.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := storage.DataDir()
			if err != nil {
				return err
			}
			if err := daemon.Install(dir); err != nil {
				return err
			}
			var pid int
			var running bool
			for range 30 { // launchd starts it a moment later
				if pid, running = daemon.Running(dir); running {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
			if !running {
				return fmt.Errorf("the daemon was installed but didn't start; see %s", ui.Path(daemon.LogPath(dir)))
			}
			w := ui.Writer(cmd.OutOrStdout())
			ui.Success(w, "Daemon running %s", ui.Faint.Render(fmt.Sprintf("(pid %d)", pid)))
			fmt.Fprintln(w, "  "+ui.Faint.Render("It starts again whenever you log in."))
			return nil
		},
	}
}

func daemonStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show whether the daemon is running",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := storage.DataDir()
			if err != nil {
				return err
			}
			w := ui.Writer(cmd.OutOrStdout())
			if pid, ok := daemon.Running(dir); ok {
				fmt.Fprintln(w, ui.Green.Render("●")+" Running "+ui.Faint.Render(fmt.Sprintf("(pid %d)", pid)))
			} else {
				fmt.Fprintln(w, ui.Yellow.Render("○")+" Not running "+ui.Faint.Render("· start it with termo daemon install"))
			}
			atLogin := "no"
			if daemon.Installed() {
				atLogin = "yes"
			}
			for _, row := range [][2]string{
				{"Starts at login", atLogin},
				{"Reminders", ui.Path(remind.NewStore(dir).Path())},
				{"Log", ui.Path(daemon.LogPath(dir))},
			} {
				fmt.Fprintf(w, "  %s %s\n", ui.Faint.Render(fmt.Sprintf("%-16s", row[0])), row[1])
			}
			return nil
		},
	}
}

func daemonUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Stop the daemon and don't start it at login",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := daemon.Uninstall(); err != nil {
				return err
			}
			w := ui.Writer(cmd.OutOrStdout())
			ui.Success(w, "Daemon stopped and removed")
			fmt.Fprintln(w, "  "+ui.Faint.Render("Reminders won't fire until you run termo daemon install again."))
			return nil
		},
	}
}

func daemonRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the daemon in this terminal",
		Long: `Run the daemon in the foreground and log each reminder as it fires.
Handy in a spare tmux pane, or under an init system such as systemd.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := storage.DataDir()
			if err != nil {
				return err
			}
			logger := log.New(ui.Writer(cmd.OutOrStdout()), "", log.LstdFlags)
			return daemon.Run(cmd.Context(), dir, remind.Notify, logger)
		},
	}
}

// daemonRunning reports whether reminders will actually fire.
func daemonRunning() bool {
	dir, err := storage.DataDir()
	if err != nil {
		return false
	}
	_, ok := daemon.Running(dir)
	return ok
}
