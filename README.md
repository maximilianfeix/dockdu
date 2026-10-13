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
[![Network](https://img.shields.io/badge/network-never-36B6C5?style=flat-square&labelColor=141E27)](#faq)
[![License](https://img.shields.io/badge/license-MIT-36B6C5?style=flat-square&labelColor=141E27)](LICENSE)

<a href="#install"><img src="https://img.shields.io/badge/Install-36B6C5?style=for-the-badge&labelColor=141E27" alt="Install"></a>
<a href="#what-it-shows"><img src="https://img.shields.io/badge/What_it_shows-141E27?style=for-the-badge" alt="What it shows"></a>
<a href="#docker-desktop"><img src="https://img.shields.io/badge/Docker_Desktop-141E27?style=for-the-badge" alt="Docker Desktop"></a>
<a href="#json"><img src="https://img.shields.io/badge/JSON-141E27?style=for-the-badge" alt="JSON"></a>

[Install](#install) · [What it shows](#what-it-shows) · [Docker Desktop](#docker-desktop) · [JSON](#json) · [How it measures](#how-it-measures) · [Compared](#compared) · [Limits](#limits) · [FAQ](#faq)

</div>

---

**`docker system df` says 11 GB. Your disk says 138 GB.** dockdu explains the difference: container log files that df never counts, the BuildKit cache split into the part you can free and the part that belongs to your images, orphan volumes, and all of it grouped by compose project. Every line that implies an action ends with a command for you to run. dockdu itself never removes, stops or prunes anything.

<div align="center">
<img src="docs/demo.gif" alt="dockdu run inside a container on Docker Desktop: summary with the gap to df, 7 MiB of unrotated logs in one compose service, build cache that prune would not free, and an orphan volume" width="900">
</div>

```bash
brew install maximilianfeix/tap/dockdu     # macOS and Linux
dockdu                                     # one read-only run, exit 0 when complete
```

On Docker Desktop the log files live inside the VM, so run the same thing as a container to see them:

```bash
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v /var/lib/docker:/var/lib/docker:ro \
  ghcr.io/maximilianfeix/dockdu
```

If it found the gigabytes you were looking for, a star helps the next person with a full disk find it.

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

The binary is static, has no configuration file, and makes no network connections. The container image is `FROM scratch` and contains the binary and nothing else.

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

COMPOSE PROJECTS  (total 4.5 GiB · shared images counted once: 12.8 MiB)
  PROJECT      CONTAINERS  IMAGES   SHARED IMAGES  VOLUMES   LOGS     TOTAL      NOTES
  dockdu-demo  3           0 B      +12.8 MiB (1)  96.0 MiB  7.6 MiB  103.6 MiB
  (none)       5           4.3 GiB  +12.8 MiB (1)  61.0 MiB  7.7 KiB  4.4 GiB    1 orphan volume(s)

VOLUMES  (3 · total 157.0 MiB · size unknown for 0)
  dockdu-orphan-demo   -   local   24.0 MiB   0   orphan (unreferenced, not from compose)
  → docker volume rm dockdu-orphan-demo   # only if you are sure nothing needs its data
```

- **Build cache, honestly.** Total, shared with images, private, and the daemon's own reclaimable figure, plus the one line df does not print: how much a prune would *not* free because the bytes belong to an image.
- **Logs per container.** Current file plus rotated siblings, the driver, whether rotation is bounded, and which compose service produced them. Measured with `stat` only; nothing is opened or read.
- **Compose projects as the unit.** Containers, images, volumes and logs per project, orphans flagged, images used by several projects marked and counted once.
- **Images with "unique if removed".** Size minus shared layers, only when no container uses the image and the daemon computed both numbers. Otherwise `unknown`, never `0`.
- **Volumes with a state.** Referenced, unreferenced, or unknown. Orphan only when unreferenced *and* not created by compose. The hint is one `docker volume rm <name>` per orphan; dockdu never suggests `volume prune` or `--volumes`.
- **Warnings instead of guesses.** A log directory it cannot read makes that container "unmeasured" and the total a lower bound, with exit code 2 so scripts notice.

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

`dockdu --json` prints the same report as one object: `summary`, `build_cache`, `logs`, `projects`, `images`, `volumes`, `desktop` (only when present), `warnings`. Byte values are integers; `-1` means unknown; lists are never `null`; every section has a `suggestions` array of plain strings.

```bash
dockdu --json | jq '.summary.not_shown_by_df_bytes, .build_cache.not_freed_by_prune_bytes'
dockdu --json | jq '.logs.items[] | select(.rotation == "unbounded") | {name, bytes}'
```

Flags: `--top N` rows per section in text output (default 10, 0 for all), `--data-root PATH` when the daemon's data root is mounted somewhere else, `--no-desktop`, `--version`. Exit codes: `0` report complete, `1` cannot reach the daemon, `2` report rendered with warnings.

<a id="how-it-measures"></a>

## How it measures

dockdu calls three Docker Engine API endpoints: `info`, `system/df` (verbose) and `containers/{id}/json`. Nothing else, and nothing that writes. It then:

1. takes the four SIZE rows of `docker system df` as the daemon reports them, never recomputed;
2. stats the log files under the data root (`<root>/containers/<id>/<id>-json.log*` for `json-file`, `<root>/containers/<id>/local-logs/container.log*` for `local`), following no symlinks and reading no contents;
3. subtracts the build-cache bytes that are shared with images once, because df counts them in both rows, and adds the log bytes: `total = df − shared build cache + logs`;
4. groups by the `com.docker.compose.project` label; images get their project from the containers that reference them, since pulled images carry no labels.

The reclaimable figure for build cache is the daemon's. It excludes shared records, which is why it differs from `docker buildx du`. An estimate of freed space is not a guarantee: on Docker Desktop the VM file only returns space when Desktop compacts it.

<a id="compared"></a>

## Compared

| | `docker system df -v` | lazydocker | dive | doku | dockdu |
|---|---|---|---|---|---|
| Logs per container | no | no | no | yes | **yes** |
| Build cache shared vs private | two commands | no | no | partly | **yes, with "prune would not free"** |
| Grouped by compose project | no | no | no | no | **yes** |
| Orphan volumes named | no | no | no | partly | **yes** |
| Docker Desktop VM file | no | no | no | no | **yes** |
| Form | CLI | TUI | TUI | web UI in a container | **one read-only command, JSON** |
| Deletes anything | no | on request | no | no | **never** |

<a id="limits"></a>

## Limits

- Linux hosts and Docker Desktop on macOS. Windows and WSL are not supported in v0.1.
- Log sizes need read access to the data root: run as root, or as the container image with `/var/lib/docker:ro` mounted. Without it the logs section is reported as unmeasured, not as zero.
- Only the daemon's built-in BuildKit builder is counted; named `docker buildx` builders keep their cache in their own volume, which shows up under volumes.
- Podman is not tested.
- Numbers are resource sizes as Docker reports them, with shared layers counted once. They are not disk-block counts and will not match `du` exactly.

<a id="faq"></a>

## FAQ

**Does it delete anything?** No. There is no flag for it. Every `→` line is text you can copy.

**Why does the host run on macOS show logs as unmeasured?** Because they are inside the Docker Desktop VM. Use the container one-liner; the read-only mount of `/var/lib/docker` resolves inside the VM.

**Why does `docker buildx du` show more reclaimable space?** It counts every record not in use; the daemon, and therefore `docker system df` and dockdu, excludes records shared with images. Pruning those only drops metadata.

**Does it talk to the network?** No. It uses the Docker socket and the local filesystem, nothing else.

**What about the `DOCKDU_RECORD_FIXTURE` environment variable?** A development aid: it writes a sanitised snapshot (compose labels only, registry hosts and build-step arguments removed) to the given path for the test suite, and refuses to overwrite an existing file. It never changes anything in Docker.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the rules (read-only forever, unknown is never zero, tests first) and [SPEC.md](SPEC.md) for the v0.1 scope. Security reports: [SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE) © Maximilian Feix
