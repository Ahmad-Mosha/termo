# Reminders

`termo remind` sets reminders from your terminal: once, after a while, or on a schedule. When one fires you get a macOS notification, plus a line in your terminal at your next prompt.

```
$ termo remind in 20m "Check the oven"
✓ Reminder #1 set
  Check the oven
  today 23:33 · in 20m
```

Reminders are fired by the [background daemon](../README.md#background-daemon). Install it once with `termo daemon install`.

## Create reminders

| Command | Example |
|---|---|
| `termo remind in <duration> <message>` | `termo remind in 20m "Check the oven"` |
| `termo remind at <time> <message>` | `termo remind at "tomorrow 9am" "Call the dentist"` |
| `termo remind every <schedule> <message>` | `termo remind every weekday 09:00 Standup` |

Quotes are optional: `termo remind in 20m Check the oven` works too. termo takes the longest run of words that reads as a time, and the rest is the message.

## Time formats

| Kind | Examples |
|---|---|
| Durations, for `in` | `20m`, `1h30m`, `2d`, `1w`, `"1 hour 30 minutes"` |
| Clock times, for `at` | `15:30`, `9am`, `9:30pm`, `noon`, `midnight` |
| Days, for `at` | `today 18:00`, `tomorrow 9am`, `fri 18:00`, `2026-09-20 09:00`, `2026-09-20` |
| Schedules, for `every` | `2h`, `hour`, `day 09:00`, `weekday 9am`, `weekend noon`, `mon,wed,fri 18:30` |

A clock time on its own means the next time that clock comes around, and a day on its own means 09:00 that day. A bare `9` is rejected because it could mean 9am or 9pm.

## Manage reminders

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
| `termo remind snooze <id> [10m]` | Remind me again later |
| `termo remind rm <id>...` | Delete reminders |
| `termo remind clear` | Delete the reminders that already fired |
| `termo remind test` | Send a test notification |
| `termo remind check` | Show reminders that fired since you last looked (the shell hook runs this) |

`termo r` is short for `termo remind`.

## Snoozing

`termo remind snooze <id> [duration]` reminds you again later: after 10 minutes unless you give a duration.

- A one-time reminder moves back by the snooze time, counted from when it was due, or from now if it already fired.
- A repeating reminder keeps its schedule and gets a one-time copy, so tomorrow's standup stays at 09:00.

## Notifications

When a reminder fires you get:

- a **macOS notification** in the top-right corner, even with no terminal open
- a line **in your terminal** above your next prompt, once [shell integration](../README.md#shell-integration) is set up:

  ```
  ⏰ Check the oven · today 23:33 #1
  ```

Run `termo remind test` to check that notifications get through.

**No notifications?** macOS shows them as coming from **Script Editor**. Open System Settings → Notifications → Script Editor, allow notifications, then run `termo remind test` again.

**Reminders don't fire?** Run `termo daemon status`. If the daemon isn't running, run `termo daemon install` and look at the log path it prints.

## How it works

- Reminders are kept in `~/.local/share/termo/reminders.json` (or under `$XDG_DATA_HOME`). Every read and write locks the file, so the CLI and the daemon never trip over each other, and writes are atomic, so a crash can't corrupt the file.
- The daemon checks the clock every second instead of setting long timers, because timers pause while a Mac sleeps. Reminders missed during sleep fire once when it wakes up.
- Notifications go through `osascript`. The reminder text is passed as an argument and never pasted into the script, so it can't inject AppleScript.
