package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Ahmad-Mosha/termo/internal/notify"
	"github.com/Ahmad-Mosha/termo/internal/remind"
	"github.com/Ahmad-Mosha/termo/internal/storage"
	"github.com/Ahmad-Mosha/termo/internal/ui"
	"github.com/spf13/cobra"
)

func remindCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remind",
		Aliases: []string{"r"},
		Short:   "Set reminders that notify you on time",
		Long: `Set reminders from your terminal: once, after a while, or on a schedule.
Run it with no subcommand to list your reminders.

  Times       15:30, 9am, noon, tomorrow 9am, fri 18:00, 2026-09-20 09:00
  Durations   20m, 1h30m, 2d, "1 hour 30 minutes"
  Schedules   2h, hour, day 09:00, weekday 9am, weekend noon, mon,wed,fri 18:00`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return listReminders(cmd, false, false) },
	}
	cmd.AddCommand(
		remindInCmd(), remindAtCmd(), remindEveryCmd(),
		remindListCmd(), remindEditCmd(), remindSnoozeCmd(), remindRmCmd(), remindClearCmd(),
		remindTestCmd(), remindCheckCmd(),
	)
	return cmd
}

func remindCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Show reminders that fired since you last looked",
		Long: `Show reminders that fired since you last looked, then forget them.
The shell hook from termo init runs this before every prompt.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// This runs before every prompt: stay quiet on errors, and only
			// take the write lock when there's something to show.
			store, err := openStore()
			if err != nil {
				return nil
			}
			var waiting bool
			if store.View(func(l *remind.List) error { waiting = len(l.Inbox) > 0; return nil }) != nil || !waiting {
				return nil
			}
			var fired []remind.Fired
			if store.Update(func(l *remind.List) error { fired = l.TakeInbox(); return nil }) != nil {
				return nil
			}
			w := ui.Writer(cmd.OutOrStdout())
			now := time.Now()
			for _, f := range fired {
				printFired(w, f, now)
			}
			return nil
		},
	}
}

func remindTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Send a test notification",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := remind.Notify(remind.Fired{Message: "This is how your reminders will look."}); err != nil {
				return err
			}
			w := ui.Writer(cmd.OutOrStdout())
			ui.Success(w, "Sent a test notification")
			fmt.Fprintln(w, "  "+ui.Faint.Render(notify.Hint))
			return nil
		},
	}
}

func remindInCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "in <duration> <message>",
		Short: "Remind me after a while",
		Example: `termo remind in 20m "Check the oven"
termo remind in 1h30m Call mom
termo remind in 2 days "Renew the domain"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, msg, err := splitArgs(args, remind.ParseDuration)
			if err != nil {
				return err
			}
			now := time.Now().Truncate(time.Second)
			return addReminder(cmd, remind.Reminder{Message: msg, Due: now.Add(d), Created: now})
		},
	}
}

func remindAtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "at <time> <message>",
		Short: "Remind me at a time",
		Example: `termo remind at 15:30 Standup
termo remind at "tomorrow 9am" "Call the dentist"
termo remind at "2026-09-20 09:00" "Renew the domain"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			now := time.Now().Truncate(time.Second)
			due, msg, err := splitArgs(args, func(s string) (time.Time, error) { return remind.ParseTime(s, now) })
			if err != nil {
				return err
			}
			return addReminder(cmd, remind.Reminder{Message: msg, Due: due, Created: now})
		},
	}
}

func remindEveryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "every <schedule> <message>",
		Short: "Remind me on a schedule",
		Example: `termo remind every 2h "Drink water"
termo remind every weekday 09:00 Standup
termo remind every mon,wed,fri 18:00 Gym`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repeat, msg, err := splitArgs(args, remind.ParseRepeat)
			if err != nil {
				return err
			}
			now := time.Now().Truncate(time.Second)
			return addReminder(cmd, remind.Reminder{Message: msg, Due: repeat.First(now), Repeat: &repeat, Created: now})
		},
	}
}

func remindListCmd() *cobra.Command {
	var all, asJSON bool
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List your reminders",
		Args:    cobra.NoArgs,
		RunE:    func(cmd *cobra.Command, _ []string) error { return listReminders(cmd, all, asJSON) },
	}
	cmd.Flags().BoolVarP(&all, "all", "a", false, "include reminders that already fired")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON for scripts")
	return cmd
}

func remindEditCmd() *cobra.Command {
	var message, in, at, every string
	var once bool
	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Change a reminder's message, time or schedule",
		Example: `termo remind edit 4 --in 30m
termo remind edit 4 --at "tomorrow 9am" --message "Call the bank"
termo remind edit 5 --every "weekday 09:30"
termo remind edit 5 --once`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			flags := cmd.Flags()
			switch {
			case flags.NFlag() == 0:
				return errors.New("say what to change: --message, --in, --at, --every or --once")
			case flags.Changed("in") && flags.Changed("at"):
				return errors.New("use --in or --at, not both")
			case flags.Changed("every") && once:
				return errors.New("use --every or --once, not both")
			}
			now := time.Now().Truncate(time.Second)
			edited, err := updateReminder(id, func(r *remind.Reminder) error {
				if flags.Changed("message") {
					if r.Message = strings.TrimSpace(message); r.Message == "" {
						return errors.New("the message can't be empty")
					}
				}
				if flags.Changed("every") {
					repeat, err := remind.ParseRepeat(every)
					if err != nil {
						return err
					}
					r.Repeat, r.Due, r.Done = &repeat, repeat.First(now), false
				}
				if once {
					r.Repeat = nil
				}
				if flags.Changed("in") {
					d, err := remind.ParseDuration(in)
					if err != nil {
						return err
					}
					r.Due, r.Done = now.Add(d), false
				}
				if flags.Changed("at") {
					t, err := remind.ParseTime(at, now)
					if err != nil {
						return err
					}
					r.Due, r.Done = t, false
				}
				return nil
			})
			if err != nil {
				return err
			}
			w := ui.Writer(cmd.OutOrStdout())
			ui.Success(w, "Reminder %s updated", idLabel(id))
			printDetails(w, edited, now)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&message, "message", "m", "", "new message")
	f.StringVar(&in, "in", "", "fire after this long, like 30m")
	f.StringVar(&at, "at", "", `fire at this time, like "tomorrow 9am"`)
	f.StringVar(&every, "every", "", `repeat on this schedule, like "weekday 09:00"`)
	f.BoolVar(&once, "once", false, "stop repeating")
	return cmd
}

func remindSnoozeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "snooze <id> [duration]",
		Short: "Push a reminder back (10m unless you say)",
		Example: `termo remind snooze 4
termo remind snooze 4 1h`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			d := remind.DefaultSnooze
			if len(args) == 2 {
				if d, err = remind.ParseDuration(args[1]); err != nil {
					return err
				}
			}
			now := time.Now().Truncate(time.Second)
			snoozed, err := updateReminder(id, func(r *remind.Reminder) error {
				r.Snooze(d, now)
				return nil
			})
			if err != nil {
				return err
			}
			ui.Success(ui.Writer(cmd.OutOrStdout()), "Snoozed %s until %s %s",
				idLabel(id), ui.When(snoozed.Due, now), ui.Faint.Render("("+ui.Until(snoozed.Due, now)+")"))
			return nil
		},
	}
}

func remindRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm <id>...",
		Aliases: []string{"delete"},
		Short:   "Delete reminders",
		Example: `termo remind rm 4
termo remind rm 4 5 6`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids := make([]int, len(args))
			for i, arg := range args {
				id, err := parseID(arg)
				if err != nil {
					return err
				}
				ids[i] = id
			}
			slices.Sort(ids)
			ids = slices.Compact(ids)

			store, err := openStore()
			if err != nil {
				return err
			}
			var removed []remind.Reminder
			err = store.Update(func(l *remind.List) error {
				for _, id := range ids {
					r, err := l.Remove(id)
					if err != nil {
						return err
					}
					removed = append(removed, r)
				}
				return nil
			})
			if err != nil {
				return err
			}
			w := ui.Writer(cmd.OutOrStdout())
			for _, r := range removed {
				ui.Success(w, "Deleted %s %s", idLabel(r.ID), r.Message)
			}
			return nil
		},
	}
}

func remindClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Delete reminders that already fired",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := openStore()
			if err != nil {
				return err
			}
			var n int
			if err := store.Update(func(l *remind.List) error { n = l.ClearDone(); return nil }); err != nil {
				return err
			}
			w := ui.Writer(cmd.OutOrStdout())
			if n == 0 {
				fmt.Fprintln(w, "Nothing to clear.")
				return nil
			}
			ui.Success(w, "Cleared %d done %s", n, plural(n, "reminder"))
			return nil
		},
	}
}

func listReminders(cmd *cobra.Command, all, asJSON bool) error {
	store, err := openStore()
	if err != nil {
		return err
	}
	var active, done []remind.Reminder
	err = store.View(func(l *remind.List) error {
		for _, r := range l.Reminders {
			if r.Done {
				done = append(done, r)
			} else {
				active = append(active, r)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	slices.SortFunc(active, func(a, b remind.Reminder) int { return a.Due.Compare(b.Due) })
	slices.SortFunc(done, func(a, b remind.Reminder) int { return b.FiredAt.Compare(a.FiredAt) })
	shown := active
	if all {
		shown = append(shown, done...)
	}

	if asJSON {
		if shown == nil {
			shown = []remind.Reminder{} // print [] rather than null
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(shown)
	}

	w := ui.Writer(cmd.OutOrStdout())
	if len(shown) == 0 {
		fmt.Fprintln(w, "No upcoming reminders. Try:")
		fmt.Fprintln(w, "  "+ui.Accent.Render(`termo remind in 20m "Take a break"`))
		return nil
	}
	printTable(w, shown, time.Now())
	footer := fmt.Sprintf("%d upcoming", len(active))
	if len(done) > 0 && !all {
		footer += fmt.Sprintf(" · %d done (--all to show)", len(done))
	}
	status := ui.Green.Render("●") + ui.Faint.Render(" daemon running")
	if !daemonRunning() {
		status = ui.Yellow.Render("○") + ui.Faint.Render(" daemon not running (termo daemon install)")
	}
	fmt.Fprintln(w, "\n  "+ui.Faint.Render(footer+" · ")+status)
	return nil
}

// updateReminder applies fn to one saved reminder and returns the result.
func updateReminder(id int, fn func(*remind.Reminder) error) (remind.Reminder, error) {
	store, err := openStore()
	if err != nil {
		return remind.Reminder{}, err
	}
	var updated remind.Reminder
	err = store.Update(func(l *remind.List) error {
		r, err := l.Get(id)
		if err != nil {
			return err
		}
		if err := fn(r); err != nil {
			return err
		}
		updated = *r
		return nil
	})
	return updated, err
}

func addReminder(cmd *cobra.Command, r remind.Reminder) error {
	store, err := openStore()
	if err != nil {
		return err
	}
	if err := store.Update(func(l *remind.List) error { r = l.Add(r); return nil }); err != nil {
		return err
	}
	w := ui.Writer(cmd.OutOrStdout())
	ui.Success(w, "Reminder %s set", idLabel(r.ID))
	printDetails(w, r, time.Now())
	if !daemonRunning() {
		fmt.Fprintln(w)
		ui.Warning(w, "The daemon isn't running, so this won't fire. Start it with %s", ui.Accent.Render("termo daemon install"))
	}
	return nil
}

func openStore() (*remind.Store, error) {
	dir, err := storage.DataDir()
	if err != nil {
		return nil, err
	}
	return remind.NewStore(dir), nil
}

// splitArgs separates the time part of args from the message. It takes the
// longest run of leading args that parse accepts, so both
// `tomorrow 9am Call mom` and `"tomorrow 9am" "Call mom"` work.
func splitArgs[T any](args []string, parse func(string) (T, error)) (T, string, error) {
	var zero T
	if _, err := parse(strings.Join(args, " ")); err == nil {
		return zero, "", errors.New(`add a message after the time, like "Take a break"`)
	}
	for n := len(args) - 1; n >= 1; n-- {
		if v, err := parse(strings.Join(args[:n], " ")); err == nil {
			if msg := strings.TrimSpace(strings.Join(args[n:], " ")); msg != "" {
				return v, msg, nil
			}
		}
	}
	_, err := parse(args[0])
	return zero, "", err
}

func parseID(s string) (int, error) {
	id, err := strconv.Atoi(strings.TrimPrefix(s, "#"))
	if err != nil || id < 1 {
		return 0, fmt.Errorf("can't read %q as a reminder ID (see termo remind ls)", s)
	}
	return id, nil
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
