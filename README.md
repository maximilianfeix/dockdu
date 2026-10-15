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

[Install](#install) · [What it shows](#what-it-shows) · [Docker Desktop](#docker-desktop) · [JSON](#json) · [`--disk`](#disk) · [How it measures](#how-it-measures) · [Compared](#compared) · [Trust](#trust) · [Limits](#limits) · [FAQ](#faq)

</div>

---

# Where did your Docker disk space go?

**dockdu shows what `docker system df` misses (container log files, build cache shared with images, storage per compose project) without deleting anything.**

In the demo below, `docker system df` on a disposable Docker Engine 29.9 host reports 313.6 MiB. dockdu finds **536.9 MiB of container log files that no df row contains**, and that all 140.7 MiB of build cache is shared with an image. `du` on the log files confirms the number. Every command and number is real; how it was recorded: [docs/verification/demo-2026-10-14.md](docs/verification/demo-2026-10-14.md).

<div align="center">
<img src="docs/demo.gif" alt="On a disposable Docker Engine 29.9 host: docker system df reports 313.6 MiB; dockdu v0.3.2 finds 536.9 MiB of container logs that df does not show, build cache shared with images, storage per compose project and an orphan volume; du -ch on the log files confirms 535.0M" width="900">
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

With Colima, OrbStack, rootless Docker or any other `docker context`, dockdu connects to the same daemon the `docker` CLI uses (DOCKER_HOST, then DOCKER_CONTEXT, then the current context). Rootless Docker is tested end to end: run dockdu as the same user, and it reads the data root under `~/.local/share/docker` without root.

### Why

`docker system df` is honest about what it counts, and silent about the rest. Container logs under `json-file`, the default driver, have no size limit and no rotation, and are not part of any df row. Build cache records can be *shared with images*. Docker Engine 29 counts them in the build cache row and Engine 28 does not. On overlay2, `docker builder prune` reports their size as reclaimed but frees almost nothing. On the containerd image store it freed about half in our measurements ([details](docs/verification/demo-2026-10-14.md#b-what-docker-builder-prune--af-actually-freed)). Volumes show up as a total without saying which stack they belong to. So people run `docker system prune`, get a couple of gigabytes back, and the disk is still full ([moby #47089](https://github.com/moby/moby/issues/47089), open since 2024; [moby #52539](https://github.com/moby/moby/issues/52539), 265 GB "used for images without any images").

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

One run, each section with its own suggestions. Real `dockdu --top 3` output (v0.3.2) from the demo host above; COMPOSE PROJECTS and IMAGES are left out here, the full run is in [docs/verification/demo-2026-10-14.md](docs/verification/demo-2026-10-14.md):

```text
SUMMARY
  docker system df reports    313.6 MiB
    images 140.8 MiB · containers 24.0 KiB · volumes 32.0 MiB
    build cache 140.7 MiB
  container logs on disk      536.9 MiB   not shown by df
  Docker resources in total   709.8 MiB
    = df − 140.7 MiB shared build cache + logs
    lower bound: shared build cache may keep its own copy (containerd store)
  data root filesystem        2.2 GiB     used of 223.6 GiB · 210.0 GiB free
    allocated blocks, a different measure from Docker's sizes

BUILD CACHE
  4 records · total 140.7 MiB · shared 140.7 MiB · private 12.0 KiB
  reclaimable (as df counts it) 12.0 KiB
  prune frees an unknown part of the 140.7 MiB shared with images: on the
  containerd image store a shared record can keep its own copy, and Docker does
  not report how much (on overlay2 none of it is freed). In use, never freed:
  0 B
  → docker builder prune --filter until=168h   # unused build cache older than 7 days
  → docker builder prune -a                    # all build cache not in use (next build is slower)

CONTAINER LOGS  (top 3 of 4 · total 536.9 MiB · complete)
  CONTAINER            PROJECT      DRIVER     SIZE       FILES  ROTATION        NOTE
  dockdu-demo-chatty   -            json-file  534.6 MiB  1      unbounded
  dockdu-demo-bounded  -            json-file  2.3 MiB    3      bounded 1m × 3
  dockdu-demo-web-1    dockdu-demo  json-file  0 B        1      unbounded
  → set "log-opts" in /etc/docker/daemon.json: {"log-driver":"json-file","log-opts":{"max-size":"10m","max-file":"3"}}   # new containers only
  → recreate the container with --log-opt max-size=10m --log-opt max-file=3   # or --log-driver local

VOLUMES  (2 · total 32.0 MiB · size unknown for 0)
  VOLUME              PROJECT      DRIVER  SIZE      REFS  STATE
  dockdu-demo_data    dockdu-demo  local   32.0 MiB  2     referenced
  dockdu-demo-orphan  -            local   0 B       0     orphan (unreferenced, not from compose)
  → docker volume rm dockdu-demo-orphan   # only if you are sure nothing needs its data
```

- **Build cache, honestly.** Total, shared with images, private and the daemon's own reclaimable figure. On overlay2 it adds the line df does not print: how much a prune would *not* free because the bytes belong to an image. On the containerd image store a prune can free part of the shared bytes, and Docker does not say how much, so dockdu says that instead of guessing.
- **buildx builders too.** A builder created with `docker buildx create --driver docker-container` keeps its whole cache in a volume named `buildx_buildkit_<node>_state`. df's build cache row does not include it, and on that host `docker builder prune -af` freed 0 B of it while `docker buildx prune --builder=<name>` freed 71 MiB ([measured](docs/verification/buildx-overlay2-2026-10-15.md)). dockdu lists each such builder under BUILD CACHE, marks its volume, and prints the prune command for that builder.
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

<a id="disk"></a>

## When `df -h` and Docker disagree: `--disk`

> Since v0.3.0.

Docker's numbers are logical sizes from its API. Your disk fills up in allocated blocks, and part of it may not be Docker's at all. Whenever the data root is visible (Linux as root, or the container image), dockdu prints the filesystem that holds it. When that filesystem is at least 80% full, it suggests `dockdu --disk`, which walks the data root the way `du -sx` does:
- `lstat` only: no symlink is followed and no file is opened, except the layer ID files under `image/overlay2/layerdb` (one short ID each, read through an `os.Root` so a link cannot leave the data root);
- each hard-linked inode is counted once;
- other filesystems, such as running containers' overlay mounts, are not entered.

It then reports what each top-level directory occupies and how much of the same filesystem lies outside the data root. Real output from a disposable Docker Engine 28.5 daemon after `scripts/demo-up.sh`:

```text
  data root filesystem        2.2 GiB     used of 223.6 GiB · 210.0 GiB free (allocated blocks, a different measure)

DATA ROOT ON DISK  (/var/lib/docker · allocated blocks · 806 entries · complete)
  ENTRY (top 8 of 12)  ON DISK    ENTRIES
  containers           109.4 MiB  43
  overlay2             73.1 MiB   651
  volumes              32.0 MiB   8
  buildkit             500.0 KiB  21
  ...
  data root on disk           215.6 MiB
  outside the data root       2.0 GiB     on the same filesystem, not under /var/lib/docker (approximate: includes filesystem metadata)
  4 mount(s) inside the data root not entered (overlay mounts of running containers, separately mounted volumes)
```

What the result tells you:
- **`outside the data root` is large:** the full disk is mostly not Docker's.
- **`overlay2` or `containerd` is far larger than images + containers + build cache:** the data root holds layers or content the API does not list. Leftovers from crashes, and on the containerd image store the compressed content kept next to the unpacked snapshots.

> Since v0.4.0.

With the overlay2 graph driver, `--disk` also says what each directory under `overlay2/` belongs to, from the driver's own records (`image/overlay2/layerdb`): image layers, container layers, and directories with no image or container record. When the last group is at least 64 MiB and a tenth of `overlay2/` beyond what all private build cache could explain, dockdu reports it and lists the largest of those directories. No Docker command lists them and no prune removes them. It never suggests deleting them by hand, because removing a directory the daemon still uses breaks images.

On the containerd image store, `overlay2/` holds the classic store from before the switch. Docker [hides those images and containers but keeps them on disk](https://docs.docker.com/engine/storage/containerd/), so `docker system df` reports 0 B and `docker system prune -af` frees 0 B ([measured](docs/verification/buildx-overlay2-2026-10-15.md)). dockdu says so even without `--disk`, from one directory listing:

```text
  hidden: the classic image store still records 9 image layer(s) in overlay2/.
  The containerd image store this daemon uses hides them, and no prune frees
  them. dockdu --disk measures them.
```

dockdu shows the logical sizes and the block counts side by side and never subtracts one from the other. The walk is bounded by 50 million entries and 10 minutes; on a 300,000-file tree it took 1.3–1.5 s and 16–18 MB of memory (`du -sx`: about 1 s). A directory it cannot read, or a budget that runs out, makes the result a lower bound with a warning.

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
# build cache: total, shared with images, what prune would not free (-1 = unknown, on the containerd image store)
dockdu --json | jq '.build_cache | {total_bytes, shared_bytes, reclaimable_bytes, not_freed_by_prune_bytes}'
# is the total complete, or a lower bound?
dockdu --json | jq '{complete: .summary.complete, logs: .summary.log_bytes_complete, unknown_volumes: .summary.volume_sizes_unknown, warnings}'
```

Flags: `--top N` rows per section in text output (default 10, 0 for all), `--disk` to measure the data root on disk (see above), `--data-root PATH` when the daemon's data root is mounted somewhere else, `--no-desktop`, `--version`. `filesystem` (and `filesystem.data_root_on_disk` with `--disk`) appear only when the data root is visible. Exit codes: `0` report complete, `1` cannot reach the daemon, `2` report rendered with warnings, `64` usage error.

<a id="how-it-measures"></a>

## How it measures

dockdu calls three Docker Engine API endpoints, `info`, `system/df` (verbose) and `containers/{id}/json`, plus the version ping the client does on first use. Nothing else, and nothing that writes. It then:

1. takes the four SIZE rows of `docker system df` as the daemon reports them, never recomputed;
2. stats the log files under the data root (`<root>/containers/<id>/<id>-json.log*` for `json-file`, `<root>/containers/<id>/local-logs/container.log*` for `local`), following no symlinks and reading no contents;
3. subtracts the build-cache bytes that are shared with images once, because df counts them in both rows, and adds the log bytes: `total = df − shared build cache + logs`. Daemons older than API 1.52 (Docker Engine 28 and earlier) already leave shared records out of df's build cache row; dockdu detects that and subtracts nothing. On the containerd image store (`daemon.image_store` in JSON) a shared record can keep its own copy, so the result there is a lower bound;
4. groups by the `com.docker.compose.project` label; images get their project from the containers that reference them, since pulled images carry no labels.

The reclaimable figure for build cache is the daemon's. It excludes shared records, which is why it differs from `docker buildx du`. An estimate of freed space is not a guarantee: on Docker Desktop the VM file only returns space when Desktop compacts it.

<a id="compared"></a>

## Compared

Checked against each tool's current docs or source on 2026-10-14; details and sources in [docs/audits/competitive-landscape.md](docs/audits/competitive-landscape.md).

| | `docker system df -v` | `docker buildx du` | lazydocker | Doku | Docker Desktop | dockdu |
|---|---|---|---|---|---|---|
| Log size per container | no | no | no | current file only | not listed | current + rotated files, driver, rotation limit |
| Build cache shared vs private | per-record flag | **shared and private totals** | no | per-record flag | reclaimable only | totals next to df, de-duplicated against images; on overlay2 "prune would NOT free", on the containerd store says the effect is unknown |
| Storage per compose project | no | no | compose-aware UI, no sizes | no | no | yes |
| Orphan volumes | LINKS = 0 | n/a | no flag | users per volume | "unused volumes" | unreferenced and not from compose, one `rm` hint each |
| Docker Desktop VM file real size | no | no | no | no | **yes, in Settings** | yes, plus OrbStack, next to the resource total |
| overlay2 / bind mount sizes | no | no | no | **yes** | not listed | `--disk`: overlay2 by owner (image, container, no record); classic store hidden by the containerd store |
| Named buildx builders | no | **yes**, one builder per call | no | no | not listed | docker-container builders found through their state volumes, one prune hint each |
| Can delete | no | no | yes, on request | no | yes, one click (Disk Usage extension) | never; the client refuses mutating requests |
| Form | CLI, JSON | CLI, JSON | TUI | long-running web UI, mounts `/`, port 9090 | GUI | one-shot CLI, versioned JSON |

[dive](https://github.com/wagoodman/dive) is a different tool (layer efficiency of one image) and works well next to dockdu.

<a id="trust"></a>

## Trust

- **It cannot delete anything.** Every Docker client dockdu creates carries a request hook that refuses any method other than GET/HEAD and any path other than `/_ping`, `/info`, `/system/df` and `/containers/<id>/json`, before the request leaves the process. A test against an instrumented fake Engine API fails if any remove, stop, kill, exec, create or prune call ever reaches it.
- **`:ro` on the socket does not make Docker read-only.** `-v /var/run/docker.sock:/var/run/docker.sock:ro` only stops the container from replacing the socket file; any process that can talk to the socket can still send POST and DELETE requests and is effectively root on the host. The guarantee comes from dockdu's code above, so run a binary you trust: a release with `checksums.txt`, or build it yourself with `go install`. For a filter outside dockdu, see the socket-proxy notes in [docs/audits/security-and-readonly-contract.md](docs/audits/security-and-readonly-contract.md).
- **It reads only metadata.** Log files are measured with `lstat`; their contents are never opened. With `--disk` it also reads the overlay2 layer ID files, which contain nothing but an ID. Names from the daemon are stripped of control characters and shell-quoted inside suggested commands.
- **No network, no telemetry.** The only connection is the Docker endpoint you configured.

<a id="limits"></a>

## Limits

- Linux hosts and Docker Desktop on macOS. Windows and WSL are not supported (inside a WSL2 distro running Docker Engine natively the Linux binary should work; not tested).
- Connecting to a daemon is not seeing its files. For a remote `tcp://` daemon, or Colima and OrbStack whose data root is inside a VM, log files are reported as unmeasured, never guessed. `ssh://` contexts are not supported; set `DOCKER_HOST` instead.
- Log sizes need read access to the data root: run as root, or as the container image with `/var/lib/docker:ro` mounted. Without it the logs section is reported as unmeasured, not as zero.
- buildx builders are found only with the docker-container driver on the same daemon, by the name of their state volume; the builder name is derived from the default node name `<builder>0` and is otherwise left to `docker buildx ls`. Kubernetes and remote builders are not seen.
- Podman is not tested.
- Numbers are resource sizes as Docker reports them, with shared layers counted once. They are not disk-block counts and will not match `du` exactly.
- On the overlay2 graph driver, `--disk` cannot tell a BuildKit snapshot from a leftover directory (BuildKit's own records are a BoltDB the daemon holds locked), so it compares their sum with the private build cache instead of naming each one. On the containerd image store, content the API does not attribute to an image or a cache record is not attributed either.

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
