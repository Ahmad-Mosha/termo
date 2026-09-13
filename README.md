# termo

[![CI](https://github.com/Ahmad-Mosha/termo/actions/workflows/ci.yml/badge.svg)](https://github.com/Ahmad-Mosha/termo/actions/workflows/ci.yml)

Everyday tools for your terminal, in one binary.

termo gathers the small tools you reach for every day under a single command, and gives them all the same clear, readable output. Reminders are the first tool, and more are on the way.

## Tools

| Tool | What it does | Status |
|---|---|---|
| [`termo remind`](docs/remind.md) | Reminders that reach you as a macOS notification and in your terminal | Available |
| [`termo ports`](docs/ports.md) | See what's listening on each port, which project or container owns it, and kill it | Available |
| `termo sys` | System health at a glance, with memory grouped by app | Planned |
| `termo net` | Network checks, such as making sure an API's rate limiting works | Planned |

Every command has built-in help: `termo --help`, `termo remind --help`.

## Install

termo needs Go 1.25 or later.

```sh
go install github.com/Ahmad-Mosha/termo@latest
```

Then set up the two pieces that tools share.

### Background daemon

The daemon keeps termo working when no terminal is open. That's how a reminder can fire on time. Install it once and it starts whenever you log in:

```sh
termo daemon install
```

| Command | What it does |
|---|---|
| `termo daemon install` | Start the daemon now and whenever you log in. Run it again after updating termo |
| `termo daemon status` | Show whether it's running, and where its data and log are |
| `termo daemon uninstall` | Stop it, and stop starting it at login |
| `termo daemon run` | Run it in the foreground, for a tmux pane or systemd |

<details>
<summary>On Linux</summary>

`termo daemon install` only supports macOS for now. On Linux, run the daemon as a systemd user service instead:

```ini
# ~/.config/systemd/user/termo.service
[Unit]
Description=termo daemon

[Service]
ExecStart=%h/go/bin/termo daemon run
Restart=on-failure

[Install]
WantedBy=default.target
```

```sh
systemctl --user enable --now termo
```

</details>

### Shell integration

termo can show things right in your terminal, such as a reminder that fired while you were busy. Add the hook for your shell:

| Shell | File | Line to add |
|---|---|---|
| zsh | `~/.zshrc` | `eval "$(termo init zsh)"` |
| bash | `~/.bashrc` | `eval "$(termo init bash)"` |
| fish | `~/.config/fish/config.fish` | `termo init fish \| source` |

The hook runs before each prompt and prints nothing unless there's something to show.

## Usage

Each tool is a subcommand. Here's a taste of each one; its guide covers the rest.

### Reminders

```sh
termo remind in 20m "Check the oven"
termo remind at "tomorrow 9am" "Call the dentist"
termo remind every weekday 09:00 Standup
termo remind                   # list your reminders
termo remind snooze 1 10m
```

Guide: [docs/remind.md](docs/remind.md)

### Ports

```sh
termo ports                    # what's listening, and where it's from
termo ports kill 3000
```

Guide: [docs/ports.md](docs/ports.md)

## Conventions

Every termo tool follows the same rules:

- **Readable output:** aligned tables, relative times like "in 20m", and errors that tell you how to fix the problem.
- **Script-friendly:** lists support `--json`, and colors are dropped when output is piped or `NO_COLOR` is set.
- **Local:** your data stays on your machine, in `~/.local/share/termo`.

## Platforms

termo is built for macOS first. It also runs on Linux, where notifications go through `notify-send`; see the Linux note under [Background daemon](#background-daemon).

## Development

```sh
go test ./...
go build -o termo . && ./termo --help
```

| Path | What's in it |
|---|---|
| `main.go` | Entry point |
| `internal/cli` | Every tool's commands and their output |
| `internal/remind` | The reminders tool: parsing, scheduling, storage |
| `internal/ports` | The ports tool: listing, matching to projects and docker |
| `internal/daemon` | The background daemon and its launchd setup |
| `internal/notify` | Desktop notifications for each OS |
| `internal/storage` | Shared data directory, file locks and safe writes |
| `internal/ui` | Shared colors, tables and friendly times |
| `docs` | A guide for each tool |

A new tool gets its own package under `internal/`, its commands in `internal/cli/<tool>.go`, and a guide in `docs/<tool>.md`.
