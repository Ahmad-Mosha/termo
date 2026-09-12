# termo

[![CI](https://github.com/Ahmad-Mosha/termo/actions/workflows/ci.yml/badge.svg)](https://github.com/Ahmad-Mosha/termo/actions/workflows/ci.yml)

Everyday tools for your terminal. It starts with reminders that actually reach you: as a macOS notification, and right in your terminal.

```
$ termo remind in 20m "Check the oven"
✓ Reminder #1 set
  Check the oven
  today 23:33 · in 20m

$ termo remind every weekday 09:00 Standup
✓ Reminder #2 set
  Standup
  every weekday at 09:00 · next Mon 09:00 · in 1d 10h
```

## Features

- **One-time reminders**: `at 15:30`, `at "tomorrow 9am"`, `at "2026-09-20 09:00"`
- **Reminders after a while**: `in 20m`, `in 1h30m`, `in 2 days`
- **Repeating reminders**: `every 2h`, `every weekday 09:00`, `every mon,wed,fri 18:30`
- **List, edit, snooze and delete** them, with `--json` output for scripts
- **macOS notifications** in the top-right corner, even when no terminal is open
- **Terminal notifications**: reminders that fired show up at your next prompt
- **Survives sleep**: reminders missed while your Mac slept fire once when it wakes up

## Install

termo needs Go 1.25 or later.

```sh
go install github.com/Ahmad-Mosha/termo@latest
```

Start the daemon. It sends the notifications, and starts again whenever you log in:

```sh
termo daemon install
```

To see reminders in your terminal too, add the hook for your shell:

| Shell | File | Line to add |
|---|---|---|
| zsh | `~/.zshrc` | `eval "$(termo init zsh)"` |
| bash | `~/.bashrc` | `eval "$(termo init bash)"` |
| fish | `~/.config/fish/config.fish` | `termo init fish \| source` |

From then on, a reminder that fired shows up above your next prompt:

```
⏰ Check the oven · today 23:33 #1
```

Last, check that notifications get through:

```sh
termo remind test
```

## Usage

### Create reminders

```sh
termo remind in 20m "Check the oven"
termo remind at 15:30 Standup
termo remind at "tomorrow 9am" "Call the dentist"
termo remind every 2h "Drink water"
termo remind every weekday 09:00 Standup
```

Quotes are optional: `termo remind in 20m Check the oven` works too. termo takes the longest run of words that reads as a time, and the rest is the message.

### Time formats

| Kind | Examples |
|---|---|
| Durations, for `in` | `20m`, `1h30m`, `2d`, `1w`, `"1 hour 30 minutes"` |
| Clock times, for `at` | `15:30`, `9am`, `9:30pm`, `noon`, `midnight` |
| Days, for `at` | `today 18:00`, `tomorrow 9am`, `fri 18:00`, `2026-09-20 09:00`, `2026-09-20` |
| Schedules, for `every` | `2h`, `hour`, `day 09:00`, `weekday 9am`, `weekend noon`, `mon,wed,fri 18:30` |

A clock time on its own means the next time that clock comes around, and a day on its own means 09:00 that day. A bare `9` is rejected because it could mean 9am or 9pm.

### Manage reminders

```
$ termo remind

  ID  MESSAGE           WHEN            IN      REPEATS
  #1  Check the oven    today 23:33     20m
  #4  Drink water       tomorrow 01:13  2h      every 2h
  #3  Call the dentist  tomorrow 09:00  9h 46m
  #2  Standup           Mon 09:00       1d 10h  every weekday at 09:00
  #5  Gym               Mon 18:30       1d 19h  every Mon, Wed, Fri at 18:30

  5 upcoming · ● daemon running
```

| Command | What it does |
|---|---|
| `termo remind ls [--all] [--json]` | List reminders. `--all` includes the ones that already fired |
| `termo remind edit <id> ...` | Change the time (`--in`, `--at`), the schedule (`--every`, `--once`) or the text (`-m`) |
| `termo remind snooze <id> [10m]` | Remind me again later. A repeating reminder keeps its schedule |
| `termo remind rm <id>...` | Delete reminders |
| `termo remind clear` | Delete the reminders that already fired |
| `termo remind test` | Send a test notification |
| `termo remind check` | Show reminders that fired since you last looked (the shell hook runs this) |

`termo r` is short for `termo remind`.

### The daemon

| Command | What it does |
|---|---|
| `termo daemon install` | Start the daemon now and whenever you log in. Run it again after updating termo |
| `termo daemon status` | Show whether it's running, and where the reminders and log are |
| `termo daemon uninstall` | Stop it, and stop starting it at login |
| `termo daemon run` | Run it in the foreground, for a tmux pane or systemd |

## How it works

- Reminders are kept in `~/.local/share/termo/reminders.json` (or under `$XDG_DATA_HOME`). Every read and write locks the file, so the CLI and the daemon never trip over each other, and writes are atomic, so a crash can't corrupt the file.
- The daemon is the same `termo` binary, run by launchd as a LaunchAgent (`~/Library/LaunchAgents/com.termo.daemon.plist`). It checks the clock every second instead of setting long timers, because timers pause while a Mac sleeps.
- Notifications go through `osascript`. The reminder text is passed as an argument and never pasted into the script, so it can't inject AppleScript.

## Troubleshooting

**No notifications?** macOS shows them as coming from **Script Editor**. Open System Settings → Notifications → Script Editor, allow notifications, then run `termo remind test`.

**Reminders don't fire?** Run `termo daemon status`. If the daemon isn't running, run `termo daemon install` and look at the log path it prints.

## Linux

termo also runs on Linux, where notifications go through `notify-send`. `termo daemon install` only supports macOS for now, so on Linux run the daemon as a systemd user service:

```ini
# ~/.config/systemd/user/termo.service
[Unit]
Description=termo reminders

[Service]
ExecStart=%h/go/bin/termo daemon run
Restart=on-failure

[Install]
WantedBy=default.target
```

```sh
systemctl --user enable --now termo
```

## Roadmap

- `termo ports`: what's listening on each port, which project or container owns it, and a way to kill it
- `termo sys`: system health at a glance, with memory grouped by app
- `termo net burst`: check that an API's rate limiting works

## Development

```sh
go test ./...
go build -o termo . && ./termo --help
```

| Path | What's in it |
|---|---|
| `main.go` | Entry point |
| `internal/cli` | Commands and their output |
| `internal/remind` | Reminders: parsing, scheduling, storage |
| `internal/daemon` | The background loop and launchd setup |
| `internal/notify` | Desktop notifications for each OS |
| `internal/storage` | Data directory, file locks, safe writes |
| `internal/ui` | Colors, tables and friendly times |
