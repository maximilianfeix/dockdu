---
title: "docker builder prune freed 0 B: two places Docker keeps disk space that no prune reaches"
tags: docker, devops, linux, containers
---

*Disclosure: this article was drafted with AI assistance (Claude) from measurements recorded in the dockdu repository. Every number comes from [the measurement log](https://github.com/maximilianfeix/dockdu/blob/main/docs/verification/buildx-overlay2-2026-10-15.md) and can be reproduced on a disposable `docker:dind` container.*

`docker system df` says the build cache is empty. `docker system prune -af` reports `Total reclaimed space: 0B`. `df -h` still shows the disk filling up. Here are two cases where all three statements are true at once, both measured on Docker Engine 29.9.0, with the commands to check each by hand.

## 1. The build cache of a buildx builder is a volume

`docker buildx create --driver docker-container` starts BuildKit in its own container, `buildx_buildkit_<node>`. The builder keeps its whole state, including its build cache, in a volume named `buildx_buildkit_<node>_state` ([buildx source](https://github.com/docker/buildx/blob/master/driver/docker-container/driver.go)). The node is usually the builder name plus `0`.

**Measured:** one builder `ci`, one build that added 30 MB.

```text
$ docker system df
TYPE            TOTAL     ACTIVE    SIZE      RECLAIMABLE
Images          2         1         430.9MB   73.34MB (17%)
Containers      1         1         40.96kB   0B (0%)
Local Volumes   2         1         73.31MB   114.7kB (0%)
Build Cache     0         0         0B        0B
```

The 73 MB of build cache is in the **Local Volumes** row. Then the two prune commands, with `du -sk` on the state volume after each:

| Command | State volume on disk | Prune printed |
|---|---|---|
| (before) | 72452 KiB | |
| `docker builder prune -af` | 72452 KiB | `Total: 0B` |
| `docker buildx prune --builder=ci -af` | 740 KiB | `Total: 73.36MB` |

`docker builder prune` works on the *selected* builder, which here was the daemon's own. A CI host that once ran `docker buildx create --use` and was later switched back keeps growing in a volume that `docker builder prune` never touches.

There is one more trap. `docker buildx rm --keep-state` removes the builder but keeps the volume. Nothing then refers to it, so it only shows up as an unused volume.

**Check by hand:**

```sh
docker volume ls --filter name=buildx_buildkit_      # one per docker-container node
docker system df -v | grep buildx_buildkit_          # sizes
docker buildx ls                                     # which builder each node belongs to
docker buildx du --builder <name>                    # that builder's cache, record by record
```

**Free it:** `docker buildx prune --builder <name>`. Add `--keep-storage 10gb` to keep a working set. Or set a GC policy in that builder's `buildkitd.toml`.

## 2. Switching to the containerd image store hides your old images

Docker Desktop and fresh Engine 29 installs use the containerd image store. If a daemon is switched to it, the images and containers of the classic overlay2 store are not deleted. Docker's docs say they become hidden: ["Your overlay2 data remains on disk but becomes hidden."](https://docs.docker.com/engine/storage/containerd/)

**Measured:** one data root with `alpine:3.22` and `nginx:1.29-alpine` pulled under the graph driver, then the same data root started with `--feature containerd-snapshotter=true`:

```text
$ docker images                 # (empty)
$ docker system df              # every row 0B
$ docker system prune -af
Total reclaimed space: 0B
$ sudo du -sh /var/lib/docker/overlay2
71.8M	/var/lib/docker/overlay2
```

On a real machine that number is whatever you had before the switch, often many gigabytes.

**Check by hand:**

```sh
docker info --format '{{.Driver}} {{json .DriverStatus}}'    # io.containerd.snapshotter.v1 = containerd store
sudo ls /var/lib/docker/image/overlay2/layerdb/sha256 | wc -l  # layers the classic store still records
sudo du -sh /var/lib/docker/overlay2
```

**Free it:** switch the daemon back to overlay2. The old images and containers reappear and can be removed normally. `docker save` anything you want to keep first.

## What does *not* hide space: containerd's compressed content

A common suspicion is that the containerd store's second copy of each layer (compressed blobs next to unpacked snapshots) is space `df` ignores. It is not, at least on 29.9:

```text
$ docker system df     → Images 296.5MB
content (compressed)     72.4 MB
snapshots (unpacked)    224.4 MB      sum 296.8 MB
```

The Images row includes both.

## Leftover overlay2 directories

The third classic case, [moby#47089](https://github.com/moby/moby/issues/47089): `overlay2/` is far larger than images, containers and build cache together. The graph driver records which directory belongs to which image layer (`image/overlay2/layerdb/sha256/*/cache-id`) and to which container (`layerdb/mounts/*/mount-id`). A directory that appears in neither is a BuildKit snapshot or a leftover. Comparing the size of those directories with the private build cache shows whether leftovers are the likely cause. Don't delete them by hand while the daemon runs.

## Tooling

[dockdu](https://github.com/maximilianfeix/dockdu), a read-only CLI that explains Docker disk usage, does all three checks since 0.4.0:
- buildx builders are listed under BUILD CACHE, with the prune command for each;
- the hidden classic store is reported from a single directory listing;
- `dockdu --disk` splits `overlay2/` into image layers, container layers and directories with no record.

It never deletes anything, and its Docker client refuses any request other than GET/HEAD on four read endpoints. The manual commands above work without it.
