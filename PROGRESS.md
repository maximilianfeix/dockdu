# Progress

A fresh session continues from here. Read SPEC.md and CLAUDE.md first.

## Current phase

Phase 7 (release candidate) complete on branch `phase-7-release` (PR #13, CI green incl. integration and CodeQL). All nine SPEC §9 criteria pass locally; evidence in `docs/verification/acceptance-2026-10-13.md`. Nothing is tagged or published.

## Done

- 2026-10-13: Repository created (github.com/maximilianfeix/dockdu), spec and rules committed on `main`.
- 2026-10-13: Toolchain installed: Go 1.27.2, GoReleaser 2.18.3. Docker Desktop (engine 29.2.0, API 1.53) running.
- 2026-10-13: Phase 0 spike answered in `docs/SPIKE.md` (Q1–Q5 plus `local` driver and Desktop VM file).
- 2026-10-13: Phase 1 skeleton: go.mod (`go 1.26`), `cmd/dockdu` with `--help`/`--version` and tests, Makefile (lint fails on unformatted files), .gitignore, MIT LICENSE, README stub, CI workflow. `make check` green locally.
- 2026-10-13: Phase 2: `internal/engine` (Snapshot model, Moby implementation using Info + DiskUsage + ContainerInspect only, Static engine, fixture load/write with no-overwrite, Sanitize), `internal/report/model.go`, `internal/collect/usage.go` (df totals and summary), fixture `internal/collect/testdata/desktop-29.2.json` recorded from this machine via `DOCKDU_RECORD_FIXTURE`. Tests green.
- 2026-10-13: Phase 3: `internal/collect/buildcache.go` (engine reclaimable, shared, private, in-use, not-freed-by-prune, duplicate IDs counted once, total mismatch warned), `images.go` (dangling, unique bytes only when Containers==0 and SharedSize known, sorted by size), `internal/report/suggest.go` (no volume prune, ever). Tests written first; 18 tests green.
- 2026-10-13: Phase 4: `internal/collect/logs.go` (json-file path from inspect or derived, local driver under local-logs/, rotated siblings summed, Lstat only, symlinks ignored, unreadable → unmeasured + incomplete, other drivers "not on disk"), 8 unit tests, integration test SPEC §9.4 creates its own `dockdu.test=1` container via the SDK and removes it by ID. `make integration` runs it inside golang:1.27 with the socket and `/var/lib/docker:ro` mounted: PASS, 3116 bytes in 3 files observed. On the macOS host the test skips (data root inside the VM).
- 2026-10-13: Phase 5: `compose.go` (projects from the compose label, images attributed through container references, images used by several projects reported per project in SharedImageBytes and counted once at section level so project totals stay additive, dangling and unused images in the (none) bucket, stopped containers older than 7 days), `volumes.go` (referenced / unreferenced / unknown; orphan only when unreferenced and not compose-created; one `docker volume rm` hint per orphan, never bulk), `AttachImageProjects`. Tests first; all green.
- 2026-10-13: Phase 6: text and JSON renderers (golden test), report builder, Docker Desktop section, CLI flags `--json --top --data-root --no-desktop`, exit codes 0/1/2/64. Collector review (phases 3–5) and render review applied: IDs validated, paths under the data root, shared images counted once at section level, unmeasured logs as -1 and counted per project, lower-bound markers, control characters neutralised, effective data root reported.

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

- Containers with label `dockdu.test=1`: `dockdu-test-logs` (1f1412f2177f), `dockdu-test-default` (871547ac7ec5), `dockdu-test-local` (34b1d5c11f04).
- Demo compose project `dockdu-demo` (containers 02ebd10c9e0b web, 01db9eef1699 worker, 653a0ccb8469 db; volume `dockdu-demo_pgdata`; compose file in the session scratchpad), labelled `dockdu.test=1`.
- Orphan demo volume `dockdu-orphan-demo` (label `dockdu.test=1`).
- Buildx builder `dockdu-builder` (container `buildx_buildkit_dockdu-builder0`, volume `buildx_buildkit_dockdu-builder0_state`).
- Local snapshot images `ghcr.io/maximilianfeix/dockdu:*` (never pushed).
- Image `alpine:3.22` (kept for the integration test).
Remove containers and volumes by ID/name after checking the label; never prune.

## Next executable step

Release (needs the owner): add the repository secret `HOMEBREW_TAP_GITHUB_TOKEN` (contents:write on maximilianfeix/homebrew-tap), then `git tag v0.1.0 && git push origin v0.1.0`; the release workflow builds archives, the cask and the ghcr image and opens a draft GitHub release to publish by hand. After that: `brew install maximilianfeix/tap/dockdu` and `docker run … ghcr.io/maximilianfeix/dockdu` are the two acceptance checks that could not run before publishing. Optional: upload `docs/social-preview.svg` (as PNG) under Settings → Social preview.
