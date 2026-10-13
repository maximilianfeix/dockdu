# Progress

A fresh session continues from here. Read SPEC.md and CLAUDE.md first.

## Current phase

Phase 2 (engine wrapper, usage collector, report model) complete on branch `phase-2-engine`. Next: Phase 3 (build cache and images).

## Done

- 2026-10-13: Repository created (github.com/maximilianfeix/dockdu), spec and rules committed on `main`.
- 2026-10-13: Toolchain installed: Go 1.27.2, GoReleaser 2.18.3. Docker Desktop (engine 29.2.0, API 1.53) running.
- 2026-10-13: Phase 0 spike answered in `docs/SPIKE.md` (Q1–Q5 plus `local` driver and Desktop VM file).
- 2026-10-13: Phase 1 skeleton: go.mod (`go 1.26`), `cmd/dockdu` with `--help`/`--version` and tests, Makefile (lint fails on unformatted files), .gitignore, MIT LICENSE, README stub, CI workflow. `make check` green locally.
- 2026-10-13: Phase 2: `internal/engine` (Snapshot model, Moby implementation using Info + DiskUsage + ContainerInspect only, Static engine, fixture load/write with no-overwrite, Sanitize), `internal/report/model.go`, `internal/collect/usage.go` (df totals and summary), fixture `internal/collect/testdata/desktop-29.2.json` recorded from this machine via `DOCKDU_RECORD_FIXTURE`. Tests green.

## Verified results (observed at runtime unless noted)

- Read-only bind mount of `/var/lib/docker` inside a container on Docker Desktop exposes `containers/<id>/*-json.log` with rotated siblings and `local-logs/container.log`; buildkit dir present.
- `local` driver leaves `LogPath` empty; file lives at `containers/<id>/local-logs/container.log`.
- Per-container `--log-opt` values appear in `HostConfig.LogConfig.Config`; daemon-level inheritance confirmed from moby source only (NOT RUN at runtime, no daemon log-opts on this machine).
- `go doc` on moby/moby/client v0.6.2: `DiskUsage`, `DiskUsageOptions{Containers,Images,BuildCache,Volumes,Verbose}`, `BuildCacheDiskUsage.Items []build.CacheRecord`, `ContainerInspect(ctx,id,ContainerInspectOptions{Size})`, `Info(ctx,InfoOptions{})`, `New`, `FromEnv`, `IsErrConnectionFailed`.
- Engine `Reclaimable` for build cache excludes shared records; `docker buildx du` does not. Spec 6.4 amended to use the engine value.
- GoReleaser 2.18.3 rejects `install_steps`; `hooks.post.install` string form parses. Spec 10 amended.
- Docker.raw: apparent 245,106,737,152 B, allocated 9,326,448 × 512 B ≈ 4.45 GiB, matches `df` inside the VM.
- With the containerd snapshotter, image bytes are not under `/var/lib/docker`; Spec 6.10 amended to compare allocated file size with API-reported sizes as separate measures.

## Open problems and blockers

- None blocking. Daemon-level log-opts inheritance remains source-verified only.
- Visual identity (banner, mark, palette) is being produced outside the repo and lands with the README in Phase 7.

## Decided spec deviations

See `docs/SPIKE.md`, section "Spec amendments proposed in this PR" (6.1, 6.4, 6.10, 10). All applied in the Phase 0 PR.

## Test resources owned by this project

Containers with label `dockdu.test=1`: `dockdu-test-logs` (1f1412f2177f), `dockdu-test-default` (871547ac7ec5), `dockdu-test-local` (34b1d5c11f04). Image `alpine:3.22`. Remove containers by ID at project end; never prune.

## Next executable step

Phase 3 on branch `phase-3-buildcache-images`: tests first for SPEC §6.4 (engine reclaimable, shared, private, not-freed-by-prune, duplicates, unknown values, empty) and §6.7 (dangling, unique bytes, -1 handling), then `internal/collect/buildcache.go`, `images.go`, `internal/report/suggest.go`.
