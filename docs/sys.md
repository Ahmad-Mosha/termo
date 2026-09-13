# System

`termo sys` shows your machine's vital signs: CPU, memory, disk, network and processes. Run it with no subcommand for a one-screen overview.

```
$ termo sys
  CPU      11.6%  ██░░░░░░░░░░░░░░░░░░  load 2.66 2.88 2.73 (10 cores)
  Memory   73.1%  ██████████████░░░░░░  12.6 GB / 17.2 GB used
  Disk     51.7%  ██████████░░░░░░░░░░  255.5 GB / 494.4 GB  /
  Uptime  1d 9h
```

## Commands

| Command | What it does |
|---|---|
| `termo sys` | A one-screen overview: CPU, memory, disk, uptime |
| `termo sys cpu` | Overall and per-core CPU usage, and load average |
| `termo sys mem` | Memory usage, grouped by app |
| `termo sys disk` | Disk usage (alias `df`) |
| `termo sys net` | Network throughput per interface |
| `termo sys ps` | Running processes (alias `top`), `--mem` to sort by memory instead of CPU |

Every command takes `--json` for scripts, and `mem`/`ps` take `-n` to change how many rows to show (10 by default).

## Memory, grouped by app

Plain process listings show a browser's helper processes as dozens of separate rows: `Chrome Helper (Renderer)`, `Chrome Helper (GPU)`, and so on. `termo sys mem` groups them under the app they belong to:

```
$ termo sys mem

  Used       73.1%  ██████████████░░░░░░  12.6 GB / 17.2 GB
  Available  4.6 GB
  Swap       1.2 GB / 2.1 GB

  APP      MEMORY    PROCS
  Browser  4.1 GB    23
  Code     1.3 GB    16
  Claude   1.3 GB    13
  Discord  567.4 MB  6
```

This groups anything named `<App> Helper`, `<App> Helper (Renderer)`, `<App> Helper (GPU)`, and so on — the pattern Chromium and Electron apps use on macOS — under `<App>`.

## Processes

`termo sys ps` samples every process's CPU usage over a short window, so it reflects what's happening right now rather than a lifetime average since the process started (which is what plain `ps %cpu` shows).

```
$ termo sys ps -n 5

  PID    NAME                       CPU    MEM       USER
  88114  Browser Helper (Renderer)  48.0%  866.4 MB  ahmadgamal
  1575   corespeechd                9.7%   32.1 MB   ahmadgamal
  13382  Browser Helper             6.7%   114.3 MB  ahmadgamal
  13365  Dia                        2.3%   588.0 MB  ahmadgamal
  22680  Browser Helper (Renderer)  1.8%   557.9 MB  ahmadgamal
```

A `—` in the MEM column means termo couldn't read that process's memory — usually because it belongs to another user, and macOS won't hand that over without root.

## Notes

- CPU and network numbers are sampled over a short window (a few hundred milliseconds) rather than read instantly, so they reflect what's happening right now instead of an average since boot.
- On macOS, disk usage is reported for `/` only. macOS mounts several synthetic system volumes (`Preboot`, `VM`, `Update`, `Data`, ...) that all share one physical container and report that whole container's usage; showing all of them would just repeat the same numbers under different names.
- `termo sys net` leaves out the loopback interface and anything that's never sent or received a byte, so only interfaces actually doing something show up.
