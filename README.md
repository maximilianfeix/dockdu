<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/banner-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="docs/banner-light.svg">
  <img src="docs/banner-dark.svg" alt="dockdu – find out where Docker's disk space actually went, and what prune won't give back." width="100%">
</picture>

[![ci](https://github.com/maximilianfeix/dockdu/actions/workflows/ci.yml/badge.svg)](https://github.com/maximilianfeix/dockdu/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/maximilianfeix/dockdu?style=flat-square&color=36B6C5&labelColor=141E27)](https://github.com/maximilianfeix/dockdu/releases/latest)
[![Go](https://img.shields.io/badge/go-1.26%2B-36B6C5?style=flat-square&labelColor=141E27)](go.mod)
[![Direct dependencies](https://img.shields.io/badge/direct_dependencies-2-36B6C5?style=flat-square&labelColor=141E27)](go.mod)
[![Read-only](https://img.shields.io/badge/docker_calls-read--only-36B6C5?style=flat-square&labelColor=141E27)](#how-it-measures)
[![Telemetry](https://img.shields.io/badge/telemetry-none-36B6C5?style=flat-square&labelColor=141E27)](#trust)
[![License](https://img.shields.io/badge/license-MIT-36B6C5?style=flat-square&labelColor=141E27)](LICENSE)

<a href="#install"><img src="https://img.shields.io/badge/Install-36B6C5?style=for-the-badge&labelColor=141E27" alt="Install"></a>
<a href="#what-it-shows"><img src="https://img.shields.io/badge/What_it_shows-141E27?style=for-the-badge" alt="What it shows"></a>
<a href="#docker-desktop"><img src="https://img.shields.io/badge/Docker_Desktop-141E27?style=for-the-badge" alt="Docker Desktop"></a>
<a href="#json"><img src="https://img.shields.io/badge/JSON-141E27?style=for-the-badge" alt="JSON"></a>

[Install](#install) · [What it shows](#what-it-shows) · [Docker Desktop](#docker-desktop) · [JSON](#json) · [How it measures](#how-it-measures) · [Compared](#compared) · [Trust](#trust) · [Limits](#limits) · [FAQ](#faq)

</div>

---

# Where did your Docker disk space go?

**dockdu shows what `docker system df` misses, without deleting anything.**

`docker system df` says 11 GB, your disk says 138 GB. dockdu explains the part of that gap Docker can account for: container log files that df never counts, the BuildKit cache split into the part you can free and the part that belongs to your images, orphan volumes, and all of it grouped by compose project. Every line that implies an action ends with a command for you to run. dockdu itself never removes, stops or prunes anything, and its Docker client refuses any request that could (see [Trust](#trust)).

<div align="center">
<img src="docs/demo.gif" alt="dockdu run inside a container on Docker Desktop: summary with the gap to df, 7 MiB of unrotated logs in one compose service, build cache that prune would not free, and an orphan volume" width="900">
</div>

```bash
brew install maximilianfeix/tap/dockdu     # macOS and Linux
dockdu                                     # one read-only run, exit 0 when complete
```

On Docker Desktop the log files live inside the VM, so a host run ends with exit code 2 and lists the logs as unmeasured. That is expected, not an error; run the same thing as a container to see them:

```bash
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v /var/lib/docker:/var/lib/docker:ro \
  ghcr.io/maximilianfeix/dockdu
```

With Colima, OrbStack, rootless Docker or any other `docker context`, dockdu connects to the same daemon the `docker` CLI uses (DOCKER_HOST, then DOCKER_CONTEXT, then the current context).

### Why

`docker system df` is honest about what it counts, and silent about the rest. Container logs under `json-file`, the default driver, have no size limit and no rotation and are not part of any df row. The build cache row does include records that are *shared with images*, which is why `docker builder prune` frees less than the number suggests. Volumes show up as a total without saying which stack they belong to. So people run `docker system prune`, get a couple of gigabytes back, and the disk is still full ([moby #47089](https://github.com/moby/moby/issues/47089), open since 2024; [moby #52539](https://github.com/moby/moby/issues/52539), 265 GB "used for images without any images").

dockdu reads the same three API calls `docker` uses, stats the log files, does the arithmetic once, and tells you what you would get back and what you would not.

<a id="install"></a>

## Install

```bash
# Homebrew (macOS, Linux)
brew install maximilianfeix/tap/dockdu

# Go
go install github.com/maximilianfeix/dockdu/cmd/dockdu@latest

# Container (any host with Docker; needed on Docker Desktop to measure logs)
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro -v /var/lib/docker:/var/lib/docker:ro ghcr.io/maximilianfeix/dockdu

# Binary: pick your platform under Releases, verify with checksums.txt
```

The binary is static, has no configuration file, and connects to nothing but the Docker endpoint. The container image is `FROM scratch` and contains the binary and nothing else.

<a id="what-it-shows"></a>

## What it shows

One run, six sections, each with its own suggestions:

```text
SUMMARY
  docker system df reports    6.5 GiB     images 4.9 GiB · containers 96.0 KiB · volumes 157.0 MiB · build cache 1.4 GiB
  container logs on disk      7.7 MiB     not shown by df
  Docker resources in total   6.1 GiB     = df − 407.5 MiB build cache shared with images + logs

BUILD CACHE
  26 records · total 1.4 GiB · shared 407.5 MiB · private 1.0 GiB · reclaimable (as df counts it) 1.0 GiB
  prune would NOT free 407.5 MiB (shared with images or in use)
  → docker builder prune --filter until=168h   # unused build cache older than 7 days

CONTAINER LOGS  (top 6 of 8 · total 7.7 MiB · complete)
  CONTAINER             PROJECT      DRIVER     SIZE       FILES  ROTATION
  dockdu-demo-web-1     dockdu-demo  json-file  7.2 MiB    1      unbounded
  dockdu-demo-worker-1  dockdu-demo  json-file  478.4 KiB  3      bounded 200k × 3
  → set "log-opts" in /etc/docker/daemon.json: {"log-driver":"json-file","log-opts":{"max-size":"10m","max-file":"3"}}

COMPOSE PROJECTS  (total 4.5 GiB · images used by several projects counted once: 12.8 MiB)
  PROJECT      CONTAINERS  IMAGES   SHARED IMAGES  VOLUMES   LOGS     TOTAL      NOTES
  dockdu-demo  3           0 B      +12.8 MiB (1)  96.0 MiB  7.6 MiB  103.6 MiB
  (none)       5           4.3 GiB  +12.8 MiB (1)  61.0 MiB  7.7 KiB  4.4 GiB    1 orphan volume(s)

VOLUMES  (3 · total 157.0 MiB · size unknown for 0)
  dockdu-orphan-demo   -   local   24.0 MiB   0   orphan (unreferenced, not from compose)
  → docker volume rm dockdu-orphan-demo   # only if you are sure nothing needs its data
```

- **Build cache, honestly.** Total, shared with images, private, and the daemon's own reclaimable figure, plus the one line df does not print: how much a prune would *not* free because the bytes belong to an image.
- **Logs per container.** Current file plus rotated siblings, the driver, whether rotation is bounded, and which compose service produced them. Measured with `stat` only; nothing is opened or read.
- **Compose projects as the unit.** Containers, images, volumes and logs per project, orphans flagged, images used by several projects marked and counted once. Image sizes include base layers shared between different images; when that makes the columns add up to more than df's image row, dockdu says by how much.
- **Images with "unique if removed".** Size minus shared layers, only when no container uses the image and the daemon computed both numbers. Otherwise `unknown`, never `0`.
- **Volumes with a state.** Referenced, unreferenced, or unknown. Orphan only when unreferenced *and* not created by compose. The hint is one `docker volume rm <name>` per orphan; dockdu never suggests `volume prune` or `--volumes`.
- **Warnings instead of guesses.** A log directory it cannot read, or a volume whose size the driver does not report, makes the total a lower bound, says so, and exits 2 so scripts notice.

<a id="docker-desktop"></a>

## Docker Desktop

On macOS, Docker Desktop keeps everything in one file, `Docker.raw`, that grows and does not shrink on its own. dockdu shows both numbers for it, apparent size (the configured maximum, which is what Finder reports) and allocated size (what is really on disk), next to the Docker-reported resource total, and says why the file does not shrink. OrbStack's `data.img` is handled the same way.

```text
DOCKER DESKTOP
  ~/Library/Containers/com.docker.docker/Data/vms/0/data/Docker.raw
  apparent size 228.3 GiB (the configured maximum) · allocated on disk 5.7 GiB
  Docker reports 6.1 GiB of resources; a different measure that may overlap, not a disk-block count
```

The log files are inside that VM, so the host binary reports them as unmeasured. The container one-liner under [Install](#install) mounts `/var/lib/docker` read-only from inside the VM and measures them.

<a id="json"></a>

## JSON

`dockdu --json` prints the same report as one object with `schema_version` (currently `1`; it changes only when a field is removed or changes meaning), `summary`, `build_cache`, `logs`, `projects`, `images`, `volumes`, `desktop` (only when present), `warnings`. Byte values are integers; `-1` means unknown; lists are never `null`; every section has a `suggestions` array of plain strings.

```bash
# containers whose logs grow without limit, largest first
dockdu --json | jq '.logs.items[] | select(.rotation == "unbounded") | {name, project, bytes}'
# the three largest compose projects
dockdu --json | jq '[.projects.items[] | select(.name != "(none)")] | sort_by(-.total_bytes)[:3] | .[] | {name, total_bytes}'
# orphan volumes (unreferenced and not created by compose)
dockdu --json | jq -r '.volumes.items[] | select(.orphan) | .name'
# build cache: total, shared with images, what prune would not free
dockdu --json | jq '.build_cache | {total_bytes, shared_bytes, reclaimable_bytes, not_freed_by_prune_bytes}'
# is the total complete, or a lower bound?
dockdu --json | jq '{complete: .summary.complete, logs: .summary.log_bytes_complete, unknown_volumes: .summary.volume_sizes_unknown, warnings}'
```

Flags: `--top N` rows per section in text output (default 10, 0 for all), `--data-root PATH` when the daemon's data root is mounted somewhere else, `--no-desktop`, `--version`. Exit codes: `0` report complete, `1` cannot reach the daemon, `2` report rendered with warnings, `64` usage error.

<a id="how-it-measures"></a>

## How it measures

dockdu calls three Docker Engine API endpoints, `info`, `system/df` (verbose) and `containers/{id}/json`, plus the version ping the client does on first use. Nothing else, and nothing that writes. It then:

1. takes the four SIZE rows of `docker system df` as the daemon reports them, never recomputed;
2. stats the log files under the data root (`<root>/containers/<id>/<id>-json.log*` for `json-file`, `<root>/containers/<id>/local-logs/container.log*` for `local`), following no symlinks and reading no contents;
3. subtracts the build-cache bytes that are shared with images once, because df counts them in both rows, and adds the log bytes: `total = df − shared build cache + logs`. Daemons older than API 1.52 (Docker Engine 28 and earlier) already leave shared records out of df's build cache row; dockdu detects that and subtracts nothing;
4. groups by the `com.docker.compose.project` label; images get their project from the containers that reference them, since pulled images carry no labels.

The reclaimable figure for build cache is the daemon's. It excludes shared records, which is why it differs from `docker buildx du`. An estimate of freed space is not a guarantee: on Docker Desktop the VM file only returns space when Desktop compacts it.

<a id="compared"></a>

## Compared

Checked against each tool's current docs or source on 2026-10-14; details and sources in [docs/audits/competitive-landscape.md](docs/audits/competitive-landscape.md).

| | `docker system df -v` | `docker buildx du` | lazydocker | Doku | Docker Desktop | dockdu |
|---|---|---|---|---|---|---|
| Log size per container | no | no | no | current file only | not listed | current + rotated files, driver, rotation limit |
| Build cache shared vs private | per-record flag | **shared and private totals** | no | per-record flag | reclaimable only | totals next to df, de-duplicated against images, "prune would NOT free" |
| Storage per compose project | no | no | compose-aware UI, no sizes | no | no | yes |
| Orphan volumes | LINKS = 0 | n/a | no flag | users per volume | "unused volumes" | unreferenced and not from compose, one `rm` hint each |
| Docker Desktop VM file real size | no | no | no | no | **yes, in Settings** | yes, plus OrbStack, next to the resource total |
| overlay2 / bind mount sizes | no | no | no | **yes** | not listed | no |
| Named buildx builders | no | **yes** | no | no | not listed | no (default builder only) |
| Can delete | no | no | yes, on request | no | yes, one click (Disk Usage extension) | never; the client refuses mutating requests |
| Form | CLI, JSON | CLI, JSON | TUI | long-running web UI, mounts `/`, port 9090 | GUI | one-shot CLI, versioned JSON |

[dive](https://github.com/wagoodman/dive) is a different tool (layer efficiency of one image) and works well next to dockdu.

<a id="trust"></a>

## Trust

- **It cannot delete anything.** Every Docker client dockdu creates carries a request hook that refuses any method other than GET/HEAD and any path other than `/_ping`, `/info`, `/system/df` and `/containers/<id>/json`, before the request leaves the process. A test against an instrumented fake Engine API fails if any remove, stop, kill, exec, create or prune call ever reaches it.
- **`:ro` on the socket does not make Docker read-only.** `-v /var/run/docker.sock:/var/run/docker.sock:ro` only stops the container from replacing the socket file; any process that can talk to the socket can still send POST and DELETE requests and is effectively root on the host. The guarantee comes from dockdu's code above, so run a binary you trust: a release with `checksums.txt`, or build it yourself with `go install`. For a filter outside dockdu, see the socket-proxy notes in [docs/audits/security-and-readonly-contract.md](docs/audits/security-and-readonly-contract.md).
- **It reads only metadata.** Log files are measured with `lstat`; their contents are never opened. Names from the daemon are stripped of control characters and shell-quoted inside suggested commands.
- **No network, no telemetry.** The only connection is the Docker endpoint you configured.

<a id="limits"></a>

## Limits

- Linux hosts and Docker Desktop on macOS. Windows and WSL are not supported (inside a WSL2 distro running Docker Engine natively the Linux binary works; not tested).
- Connecting to a daemon is not seeing its files. For a remote `tcp://` daemon, or Colima and OrbStack whose data root is inside a VM, log files are reported as unmeasured, never guessed. `ssh://` contexts are not supported; set `DOCKER_HOST` instead.
- Log sizes need read access to the data root: run as root, or as the container image with `/var/lib/docker:ro` mounted. Without it the logs section is reported as unmeasured, not as zero.
- Only the daemon's built-in BuildKit builder is counted; named `docker buildx` builders keep their cache in their own volume, which shows up under volumes.
- Podman is not tested.
- Numbers are resource sizes as Docker reports them, with shared layers counted once. They are not disk-block counts and will not match `du` exactly.
- Disk space Docker's API does not report is not attributed: overlay2 directories left behind by crashes, and with the containerd image store the compressed content kept next to unpacked snapshots. If dockdu's total is far below what `du -sx /var/lib/docker /var/lib/containerd` shows, that is where the rest is.

<a id="faq"></a>

## FAQ

**Does it delete anything?** No. There is no flag for it, and the client refuses mutating requests (see [Trust](#trust)). Every `→` line is text you can copy.

**Why does the host run on macOS show logs as unmeasured?** Because they are inside the Docker Desktop VM. Use the container one-liner; the read-only mount of `/var/lib/docker` resolves inside the VM.

**Why does `docker buildx du` show more reclaimable space?** It counts every record not in use; the daemon, and therefore `docker system df` and dockdu, excludes records shared with images. Pruning those only drops metadata.

**Does it talk to the network?** Only to the Docker endpoint you configured (normally a local socket). Nothing else.

**What about the `DOCKDU_RECORD_FIXTURE` environment variable?** A development aid: it writes a sanitised snapshot (compose labels only, registry hosts and build-step arguments removed) to the given path for the test suite, and refuses to overwrite an existing file. It never changes anything in Docker.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the rules (read-only forever, unknown is never zero, tests first) and [SPEC.md](SPEC.md) for the v0.1 scope. Security reports: [SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE) © Maximilian Feix
