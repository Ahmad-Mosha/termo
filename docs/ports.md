# Ports

`termo ports` shows what's listening on your machine: which process owns each port, which project or docker container it belongs to, and how long it's been up. `termo ports kill` stops it.

```
$ termo ports

PORT   PROCESS   PID    WHERE                            UP
5432   postgres  3036   /opt/homebrew (stable)           1d 8h
18080  caddy     —      docker: termo-test               1 second
18123  Python    82334  ~/code/termo (ports-cli)          2s
```

## WHERE

termo figures out what's behind a port in this order:

1. **A docker container**, if one publishes that port. termo asks `docker ps` directly, so this works with Docker Desktop, OrbStack, Colima, or any setup where the `docker` CLI talks to a running daemon. On macOS the port's real owner is Docker's own backend process, not the container — termo shows the container instead, since that's what you actually care about.
2. **A git project**, if the process's working directory is inside one. termo walks up from the working directory to find `.git` (following worktrees and submodules to their real `HEAD`) and shows the project root and current branch.
3. Otherwise, just the process's working directory, or `—` if that can't be read.

## Commands

| Command | What it does |
|---|---|
| `termo ports [--json]` | List what's listening. `--json` prints structured rows for scripts |
| `termo ports kill <port> [--force]` | Stop whatever is listening. `--force` skips asking nicely |

`termo p` is short for `termo ports`.

## Stopping something

```
$ termo ports kill 3000
✓ Stopped port 3000 (node)

$ termo ports kill 18081 --force
✓ Killed port 18081 (docker: termo-test)
```

By default, `kill` asks nicely: `SIGTERM` for a process, `docker stop` for a container, so it gets a chance to shut down cleanly. `--force` is immediate: `SIGKILL`, or `docker kill`.

If a container owns the port, termo always stops the container — not the process holding the socket on your host, which on macOS is Docker's own backend and not something you want to kill.

## Notes

- Reading another user's process can fail (permission denied), in which case its working directory shows as `—`. This is normal for system daemons and other users' processes.
- termo only looks at TCP ports.
- Without docker installed, or with the daemon not running, `termo ports` still works fine — it just won't have anything to show under "docker:".
