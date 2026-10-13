# dockdu — v0.1 specification

Status: authoritative for v0.1. Facts marked (verified 2026-10-13) were checked against official docs or source on that date. Anything not in this file is out of scope for v0.1.

## 1. Goal

dockdu is a read-only CLI that explains where Docker's disk space actually went, and what `docker system prune` would NOT give back. It answers the question people ask in forums when `df -h` says the disk is full but `docker system df` reports a fraction of that.

One sentence for the README: "Find out where Docker's disk space actually went, and what prune would NOT give back."

## 2. Users and scenario

- Developer on Docker Desktop for macOS whose disk filled up.
- Operator of a Linux host (homelab, VPS, CI runner) running docker-compose stacks.
- Both have already tried `docker system prune` and got back less than expected.

Before: `df -h` 92 %, `docker system df` 11 GB, prune frees 2 GB, an hour of searching.
After: one command, a few seconds, a report that attributes the gap and names a safe command per finding.

## 3. v0.1 scope (must have)

1. `dockdu` prints a human-readable report. `dockdu --json` prints the same data as JSON.
2. Report sections, in this order:
   1. Summary: what `docker system df` reports (the sum of its four SIZE rows), the bytes dockdu measured on top (container log files, labelled "not shown by df"), and a resource total = images + containers + volumes + (build cache total − build cache shared with images) + logs. Shared build-cache bytes are already inside the image sizes and are never added twice (review finding, 2026-10-13). The total counts shared layers once, as df does; it is not a disk-block count.
   2. Build cache: total, shared, private, reclaimable, and the line "prune would NOT free <X> (shared with images or in use)".
   3. Container logs: per container, log file size including rotated files, log driver, whether rotation is bounded, top N by size.
   4. Compose projects: per project (label `com.docker.compose.project`): containers, images referenced, volumes, log bytes, total. A final group "no compose project" for everything else. Orphans flagged: volumes with `RefCount == 0`, dangling images, stopped containers older than 7 days.
   5. Images: total, dangling, "unique bytes freed if removed" = `Size - SharedSize` for images with `Containers == 0`.
   6. Volumes: per volume size (or "unknown" when `-1`), reference count, compose project.
   7. Docker Desktop (macOS only, only when the VM disk file exists): apparent size vs allocated size of the VM disk file, next to the sum of Docker-reported resource sizes, with a note that the two are different measures that may overlap, and a fixed one-paragraph explanation of why the file does not shrink.
3. Every finding line that implies an action ends with one suggested command, printed as text. dockdu never executes anything that modifies state.
4. Works in two ways: as a local binary, and as a container with
   `docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro -v /var/lib/docker:/var/lib/docker:ro ghcr.io/maximilianfeix/dockdu`
5. Degrades gracefully: if a data source is unavailable (no read access to the data root, no Desktop VM file), the section prints a one-line warning and the rest of the report still renders. Exit code 2 in that case.

## 4. Non-goals for v0.1

- No deletion, no `--fix`, no prompts to run commands.
- No Windows, no WSL, no Podman, no remote hosts beyond `DOCKER_HOST`.
- No slowness, CPU, memory, OOM or restart-loop diagnostics.
- No daemon mode, no web UI, no TUI, no watch mode.
- No Kubernetes, no Swarm.
- No non-default BuildKit builders (only the daemon's built-in builder).
- No config file, no plugins, no telemetry.

## 5. CLI surface

```
dockdu [flags]

  --json              machine-readable output
  --top N             rows per section (default 10)
  --data-root PATH    override Docker data root (default: from `docker info` DockerRootDir, fallback /var/lib/docker)
  --no-desktop        skip the Docker Desktop section
  --version           print version and exit
  --help

Exit codes: 0 report complete · 1 cannot reach the Docker daemon · 2 report rendered with warnings · 64 usage error (EX_USAGE)
Honours DOCKER_HOST, DOCKER_CERT_PATH, DOCKER_TLS_VERIFY, DOCKER_API_VERSION via client.FromEnv.
```

## 6. Data sources and exact APIs (verified 2026-10-13)

### 6.1 Go modules

- Engine client: `github.com/moby/moby/client` v0.6.2 (2026-10-08). Types: `github.com/moby/moby/api` v1.56.1 (2026-10-01).
- Do NOT use `github.com/docker/docker`. It is deprecated since Engine 29.0, frozen at v28.5.2+incompatible and carries Go vulnerability advisories GO-2026-4883 and GO-2026-4887.
- Minimum Go: 1.26 (required by the SDK since 29.8.0). Pin `go 1.26` in go.mod. Installed on the development machine: go1.27.2.
- Only these two direct third-party modules in v0.1. Their transitive requirements (go-connections, go-units, distribution/reference, containerd/errdefs, opencontainers specs, go-winio, x/sys, OpenTelemetry) are accepted and listed in docs/SPIKE.md. Standard library for everything else (flag, encoding/json, text/tabwriter, os, syscall). No cobra, no color library. If a third direct module seems necessary, stop and explain why.

### 6.2 Client construction

```go
import "github.com/moby/moby/client"

cli, err := client.New(client.FromEnv) // API version negotiation is the default
```

`client.NewClientWithOpts` and `client.WithAPIVersionNegotiation()` are deprecated. Do not use them.

### 6.3 Disk usage call

```go
res, err := cli.DiskUsage(ctx, client.DiskUsageOptions{
    Containers: true, Images: true, Volumes: true, BuildCache: true, Verbose: true,
})
// res.Containers.Items []container.Summary
// res.Images.Items     []image.Summary
// res.Volumes.Items    []volume.Volume
// res.BuildCache.Items []build.CacheRecord   (type path to confirm in Phase 0)
// each aggregate also has TotalSize, Reclaimable, ActiveCount, TotalCount
```

Packages: `github.com/moby/moby/api/types/{container,image,volume,build}`.

Relevant fields:
- `image.Summary`: `ID`, `RepoTags`, `Size`, `SharedSize` (-1 when not calculated), `Containers` (-1 when not calculated), `Labels`. There is no `VirtualSize` field any more.
- `container.Summary`: `ID`, `Names`, `ImageID`, `Labels`, `State`, `Created`, `SizeRw`, `SizeRootFs`.
- `volume.Volume`: `Name`, `Driver`, `Labels`, `Mountpoint`, `UsageData *UsageData{Size, RefCount}` (Size only for the `local` driver, otherwise -1; RefCount -1 when unknown).
- `build.CacheRecord`: `ID`, `Parents`, `Type`, `Description`, `InUse`, `Shared`, `Size`, `CreatedAt`, `LastUsedAt *time.Time`, `UsageCount`.

### 6.4 Build cache arithmetic (amended after the Phase 0 spike)

- total = `BuildCache.TotalSize` from the API (equals the sum of record sizes).
- reclaimable = `BuildCache.Reclaimable` from the API. This is the number `docker system df` prints. Observed on 2026-10-13: the engine excludes shared records, so it differs from `docker buildx du`, which counts every `!InUse` record. dockdu reports the engine value and never recomputes it.
- shared = sum of Size where `Shared`
- private = total - shared
- "prune would NOT free" = sum of Size where `Shared || InUse`. Pruning a shared record only deletes metadata; the bytes stay with the image.
- An estimate of freed space is not a guarantee: physical space inside a Docker Desktop VM file is only returned when Desktop compacts it.
- Source for the engine rule: moby `daemon/internal/builder-next/builder.go`, `Builder.DiskUsage` counts a record as reclaimable only when `!InUse && !Shared`. Invariant to assert in tests: `Reclaimable + Σ Size(Shared || InUse) == TotalSize`.
- `CacheRecord.Parents` carries an odd json tag (`" Parents"`, leading space) in api v1.56.1; it did decode on engine 29.2 in the recorded fixture, but dockdu does not use the field.
- Suggested command for reclaimable private cache: `docker builder prune --filter until=168h`. For everything unused: `docker builder prune -a`.

### 6.5 Container log files

- `cli.ContainerInspect(ctx, id, client.ContainerInspectOptions{})` returns `.Container.LogPath` and `.Container.HostConfig.LogConfig{Type, Config map[string]string}`.
- `json-file` driver: `LogPath` = `<data-root>/containers/<id>/<id>-json.log`. Rotated files: `<id>-json.log.1`, `.2`, … Defaults: `max-size` unlimited, `max-file` 1. Rotation is bounded only when `Config["max-size"]` is set.
- `local` driver: the daemon leaves `LogPath` EMPTY (observed). Derive `<data-root>/containers/<id>/local-logs/container.log` (observed); rotated files are `container.log.1.gz` … and the defaults are 20m × 5 files, compressed (from the driver docs, not observed). dockdu sums every regular file named `container.log*` in that directory.
- Other drivers (journald, syslog, none): report "not on disk under data root" and skip sizing.
- Size = sum of the current file plus all rotated siblings, using `os.Stat`. If the directory is unreadable, warn once and continue.
- Suggested commands: unbounded json-file log → show how to set `log-opts` in `/etc/docker/daemon.json` (`"max-size": "10m", "max-file": "3"`, affects new containers only) and the per-container `--log-opt` flags. Never suggest deleting the file while the container runs.

### 6.6 Compose grouping

Label keys (from docker/compose `pkg/api/labels.go`):
`com.docker.compose.project`, `com.docker.compose.service`, `com.docker.compose.volume`, `com.docker.compose.network`, `com.docker.compose.project.working_dir`, `com.docker.compose.project.config_files`.

- Containers, volumes and networks created by compose carry the project label.
- Images carry compose labels only when compose BUILT them. Pulled images have none. Therefore attribute an image to a project via the containers that reference it (`container.Summary.ImageID` → project label). An image referenced by containers of several projects is listed under each with a "shared" marker and counted once in totals.

### 6.7 Images

- Dangling: `len(RepoTags) == 0` or `RepoTags == ["<none>:<none>"]`.
- Unique bytes freed if removed: `Size - SharedSize`, only when `Containers == 0` and both values are not -1. This mirrors the daemon's own reclaimable computation.
- Suggested commands: `docker image prune` (dangling only), `docker image prune -a` (all unused, with a warning that it removes images not referenced by any container).

### 6.8 Volumes

- Size from `UsageData.Size`, "unknown" when -1 or when UsageData is nil.
- Unreferenced when `RefCount == 0`. Orphan when unreferenced AND no compose project label.
- Suggested command: `docker volume rm <name>` listed per orphan. Never suggest `docker volume prune` or any `--volumes` flag.

### 6.9 Data root

- Read `cli.Info(ctx).DockerRootDir` first; fall back to `/var/lib/docker`; `--data-root` overrides both.

### 6.10 Docker Desktop on macOS

- VM disk file: `~/Library/Containers/com.docker.docker/Data/vms/0/data/Docker.raw`. Apparent size = `os.Stat().Size()`. Allocated size = `syscall.Stat_t.Blocks * 512`. Many tools show the apparent size, which is the configured maximum.
- OrbStack: `~/Library/Group Containers/HUAQ24HBR6.dev.orbstack/data/data.img`, same measurement. Label the section "OrbStack" when only this file exists.
- Fixed explanation paragraph (print verbatim, cite the Docker FAQ): space inside the VM file is freed when images, containers, volumes or build cache are removed. The file itself shrinks only when Docker Desktop compacts it; lowering the "Disk usage limit" in Settings → Resources → Advanced deletes the file and all images.
- Detection: only on `runtime.GOOS == "darwin"` and only when the file exists. Inside the container image this section is always skipped.
- Spike result 2026-10-13, observed on one machine: `/var/lib/docker/overlay2` and `image` do not exist with the containerd snapshotter and `du -sx /var/lib/docker` is tiny; where image bytes live was not determined. dockdu therefore does not measure inside the VM. It shows the allocated file size and the resource total from the summary as two separate numbers and says so.

## 7. Output format

Human-readable, width 100, aligned with `text/tabwriter`, sizes in binary units with one decimal (`6.2 GiB`). No colour in v0.1. Example (numbers illustrative):

```
dockdu v0.1.0 · docker 29.1.0 · data root /var/lib/docker

SUMMARY
  docker system df reports    84.2 GiB   images 5.0 · containers 0.1 · volumes 15.0 · build cache 64.1
  container logs on disk      48.3 GiB   not shown by df
  Docker resources in total  101.5 GiB   = df − 31.0 GiB build cache shared with images + logs

BUILD CACHE
  total 64.1 GiB · shared 31.0 GiB · private 33.1 GiB · reclaimable 33.1 GiB
  prune would NOT free 31.0 GiB (shared with images or in use)
  → docker builder prune --filter until=168h

CONTAINER LOGS  (top 10 of 23)
  CONTAINER        PROJECT   DRIVER     SIZE      ROTATION
  shop-web-1       shop      json-file  6.2 GiB   unbounded
  blog-worker-3    blog      json-file  4.1 GiB   unbounded
  …
  → set "log-opts" in /etc/docker/daemon.json: {"max-size":"10m","max-file":"3"} (new containers only)

COMPOSE PROJECTS
  PROJECT     CONTAINERS  IMAGES    VOLUMES   LOGS      TOTAL
  shop        4           3.9 GiB   12.1 GiB  6.4 GiB   22.4 GiB
  blog        3           2.1 GiB   2.8 GiB   4.3 GiB   9.2 GiB
  (none)      2           1.0 GiB   13.8 GiB  0.1 GiB   14.9 GiB   ← 3 orphan volumes, 7 dangling images

…
```

JSON: one top-level object with keys `summary`, `build_cache`, `logs`, `projects`, `images`, `volumes`, `desktop`, `warnings`. All byte values are integers in bytes. Suggested commands appear as strings in a `suggestions` array per section.

## 8. Package layout

```
cmd/dockdu/main.go            flags, exit codes, wiring only
internal/engine/              thin wrapper around the moby client behind an interface (for fakes)
internal/collect/usage.go     DiskUsage call → model
internal/collect/logs.go      inspect + file sizes
internal/collect/buildcache.go
internal/collect/compose.go   grouping
internal/collect/desktop.go   darwin only (build tag)
internal/report/model.go      the Report struct (what --json serialises)
internal/report/text.go       tabwriter renderer
internal/report/json.go
internal/report/suggest.go    static suggestion strings
internal/collect/testdata/    JSON fixtures recorded from a real daemon
```

Errors are wrapped with `%w`. No panics outside `main`. Every collector returns `(partial result, []warning, error)`.

## 9. Acceptance criteria (runnable)

1. `go build ./... && go vet ./... && go test ./...` exit 0.
2. `gofmt -l .` prints nothing.
3. Unit tests cover build-cache arithmetic, log-path derivation for json-file and local, compose grouping incl. shared images, image unique-size rule, dangling detection, and the text and JSON renderers, all from fixtures.
4. `DOCKDU_INTEGRATION=1 go test ./...` creates a throwaway container with `--log-opt max-size=1k --log-opt max-file=3`, writes more than 3 KiB of output, and asserts that the report shows the container with rotation "bounded" and a size including the rotated file. The test skips when the daemon is unreachable.
5. `./dockdu --json | jq -e '.summary.docker_bytes_total > 0'` exits 0 on a host with Docker.
6. On a host without a reachable daemon, `./dockdu` exits 1 and prints one line naming `DOCKER_HOST` and the socket it tried.
7. `docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro -v /var/lib/docker:/var/lib/docker:ro ghcr.io/maximilianfeix/dockdu` prints the same report as the local binary on the same host.
8. README contains `demo.gif` produced by `vhs demo.tape`, the one-line install commands, the "what it shows" list and the "never deletes anything" statement.
9. `goreleaser release --snapshot --clean` succeeds locally.

## 10. Release procedure

- GoReleaser v2. `builds`: linux and darwin, amd64 and arm64, `CGO_ENABLED=0`, `-trimpath`, ldflags set `main.version`.
- `homebrew_casks` (not the deprecated `brews`) publishing to `maximilianfeix/homebrew-tap`, directory `Casks`. Unsigned binary: post-install hook running `xattr -dr com.apple.quarantine` on the staged binary, written as the `hooks.post.install` string form (`system_command "/usr/bin/xattr", args: [...]`), because GoReleaser 2.18.3 does not know `install_steps` (spike Q5).
- `dockers_v2` (not the deprecated `dockers`) publishing `ghcr.io/maximilianfeix/dockdu` for linux/amd64 and linux/arm64 from a `FROM scratch` Dockerfile with `COPY $TARGETPLATFORM/dockdu /usr/bin/dockdu` and `ENTRYPOINT ["/usr/bin/dockdu"]`.
- Tag `v0.1.0`, `gh release create v0.1.0 --generate-notes`. The release workflow needs a token with `contents:write` on the tap repo and `write:packages` for ghcr.
- `.github/workflows/ci.yml`: build, vet, test, gofmt on push and PR. `.github/workflows/release.yml`: goreleaser on tag.
- License: MIT. Commits: conventional commits, no AI attribution lines.

## 11. Phase 0 spike questions (answer before Phase 1)

1. Does `-v /var/lib/docker:/var/lib/docker:ro` inside a container on Docker Desktop macOS expose `containers/<id>/*-json.log` and `buildkit/` readably? Test with `docker run --rm -v /var/lib/docker:/var/lib/docker:ro alpine ls /var/lib/docker/containers`.
2. Are daemon-level `log-opts` from `daemon.json` reflected in each container's `HostConfig.LogConfig.Config`, or left empty? Test with a daemon.json setting and `docker inspect`.
3. Confirm the concrete Go type of `DiskUsageResult.BuildCache.Items` in moby/moby/client v0.6.2 with `go doc github.com/moby/moby/client DiskUsageResult`.
4. Confirm `go doc github.com/moby/moby/client Client.DiskUsage` and `Client.ContainerInspect` signatures before writing the wrapper.
5. Confirm that `goreleaser` installed locally supports `homebrew_casks.hooks.post.install_steps`; otherwise use the older `hooks.post.install` Ruby string form.

Record the answers in `docs/SPIKE.md` and adjust this spec where they contradict it.

## 12. Follow-ups explicitly deferred

Windows and WSL; `--fix` with confirmation; Podman socket; non-default builders via the BuildKit gRPC hijack; CI mode with thresholds and non-zero exit on budget; colour output; TUI.
