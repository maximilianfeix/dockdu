# Performance

Date: 2026-10-14 · Machine: Apple Silicon Mac (darwin/arm64), Go 1.27.2, Docker Desktop with engine 29.2.0 (API 1.53).

## What dockdu costs

A run is:
- 1 version ping, 1 `GET /info` and 1 `GET /system/df?verbose=1`;
- 1 `GET /containers/<id>/json` per container;
- 1 `ReadDir` plus one `lstat` per matching file per container log directory.

No file contents are read. The daemon-side cost is dominated by `/system/df`: the daemon walks image layers, container write layers and local volumes to size them. On hosts with large volumes that call alone can take seconds, and dockdu cannot make it faster. `docker system df -v` pays the same cost.

## Measurements

| Scenario | Result | How |
|---|---|---|
| Empty daemon (0 images, 0 containers), warm | 0.01–0.02 s wall, 14–15 MB peak RSS | `/usr/bin/time -l dockdu --json`, 3 runs |
| First run of a freshly built binary on macOS | 3.9 s wall, 0.01 s CPU | same; the time is spent waiting, not in dockdu. The assumed cause is macOS assessing a new unsigned binary (not verified). |
| Fake Engine, 200 containers, 2 ms per inspect, 1 worker (v0.1 behaviour) | 533–570 ms per snapshot, 4.5 MB allocated, 38 k allocs | `go test -bench SnapshotSlowInspect ./internal/engine` |
| Same, 4 workers (this branch) | 132–147 ms per snapshot, 5.3–5.5 MB allocated, 40 k allocs | same |
| Binary size | 11.5 MB (darwin/arm64, not stripped) | `ls -l` |

The concurrency change is bounded at four in-flight inspects. Each inspect is a cheap metadata read on the daemon side; four keeps a daemon under load from being flooded while hiding socket round trips. The benchmark models latency only. It does not measure daemon CPU under real load (**NOT RUN**: no large real host was available).

## Not yet measured (BLOCKED: needs a real host)

- Linux host with 500+ containers and thousands of rotated log files: total time, time spent in `/system/df` vs inspect vs `lstat`.
- Large BuildKit cache (10k+ records): JSON decode time and memory.
- Slow or overloaded daemon: behaviour at the 90 s timeout.

## Reproduce

```
go test -run XXX -bench SnapshotSlowInspect -benchmem ./internal/engine
go build -o /tmp/dockdu ./cmd/dockdu && /usr/bin/time -l /tmp/dockdu --json > /dev/null   # macOS; use /usr/bin/time -v on Linux
```
