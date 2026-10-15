# Verification: buildx builders and overlay2 attribution (2026-10-15)

Every run below used a disposable `docker:*-dind` container labelled `dockdu.test=1` on Docker Desktop (arm64), removed afterwards with its volumes. dockdu was the `buildx-builder-volumes` branch, built for linux/arm64 and run inside the dind container with the data root visible.

## A. docker-container buildx builder (Engine 29.9.0, buildx v0.38.0)

```sh
docker buildx create --name ci --driver docker-container
docker buildx build --builder ci -t t1 --load /ctx      # FROM alpine:3.22, RUN 30 MB of random data
docker buildx create --name old --driver docker-container
docker buildx inspect --bootstrap old && docker buildx rm --keep-state old
```

What Docker reports:

```text
$ docker system df
TYPE            TOTAL     ACTIVE    SIZE      RECLAIMABLE
Images          2         1         430.9MB   73.34MB (17%)
Containers      1         1         40.96kB   0B (0%)
Local Volumes   2         1         73.31MB   114.7kB (0%)
Build Cache     0         0         0B        0B
$ docker buildx du --builder ci | tail -2
Reclaimable:	73.36MB
Total:		73.36MB
```

dockdu (excerpt):

```text
BUILD CACHE
  0 records · total 0 B · shared 0 B · private 0 B
  ...
  buildx builders with the docker-container driver keep their own cache in a
  state volume. It is counted under VOLUMES, not in the figures above:
  BUILDER  NODE  SIZE       STATE
  ci       ci0   70.1 MiB   builder container exists
  old      old0  112.0 KiB  no container: builder removed, state kept
  → docker buildx prune --builder=ci   # this builder's cache lives in its state volume, not in df's build cache row
```

What each prune frees, measured with `du -sk` on the state volume:

| Step | `du -sk …/buildx_buildkit_ci0_state` | Prune printed |
|---|---|---|
| before | 72452 KiB | |
| `docker builder prune -af` | 72452 KiB (0 freed) | `Total: 0B` |
| `docker buildx prune --builder=ci -af` | 740 KiB (71.0 MiB freed) | `Total: 73.36MB` |

`docker builder prune` acts on the selected builder, here the daemon's own; it does not reach a docker-container builder unless that builder is selected with `docker buildx use`.

## B. overlay2 by owner (Engine 28.5.2, overlay2 graph driver)

Built a two-layer image, started one container and created another. `dockdu --disk`:

```text
  overlay2 by owner
    image layers 27.9 MiB in 3 · container layers 128.0 KiB in 4
    no image or container record 28.0 KiB in 2
```

The two unrecorded directories are BuildKit snapshots (12 and 16 KiB). The first version of the feature reported their block overhead as "27.9 KiB more than private build cache could explain"; that false alarm led to the threshold (≥ 64 MiB and ≥ 10 % of overlay2). With the threshold: "within what BuildKit snapshots ... and block overhead explain".

Then a 50 MB directory without any record was planted under `overlay2/` to stand in for a leftover (the disposable daemon only):

```text
    no image or container record 47.7 MiB in 3
    47.7 MiB of these is more than all private build cache could explain
    ...
    LARGEST WITHOUT RECORD     ON DISK
    leftoverx1abc              47.7 MiB
```

This shows the detection on a planted directory, not on a real crash leftover.

## C. Classic store hidden by the containerd image store (Engine 29.9.0)

One data-root volume, first with `--feature containerd-snapshotter=false` (overlay2 graph driver), then with `=true`:

```text
# graph driver
overlay2
alpine:3.22 8.54MB
nginx:1.29-alpine 61.6MB
71.8M	/var/lib/docker/overlay2

# same data root, containerd image store
overlayfs [["driver-type","io.containerd.snapshotter.v1"]]
$ docker images          → (empty)
$ docker system df       → every row 0B
$ docker system prune -af
Total reclaimed space: 0B
$ du -sh /var/lib/docker/overlay2
71.8M	/var/lib/docker/overlay2
```

dockdu without `--disk`:

```text
  hidden: the classic image store still records 9 image layer(s) in overlay2/.
  The containerd image store this daemon uses hides them, and no prune frees
  them. dockdu --disk measures them.
  → dockdu --disk   # read-only: ...
```

dockdu with `--disk`:

```text
  overlay2: 71.8 MiB on disk from the classic image store. This daemon uses the
  containerd image store, which hides the classic store's images and containers
  but keeps them on disk: docker images, docker system df and every prune leave
  them out. Switch the daemon back to overlay2 to see and remove them; docker
  save any you still need first.
```

A fresh `docker:29-dind` (containerd image store from the start, one pull, one build) has no `overlay2/` directory, and dockdu prints neither hint.

## D. containerd image store: does df count the compressed content? (Engine 29.9.0)

Fresh `docker:29-dind`, `nginx:1.29-alpine` and `python:3.13-slim` pulled:

```text
$ docker images
IMAGE               ID             DISK USAGE   CONTENT SIZE
nginx:1.29-alpine   5616878291a2       92.7MB         26.7MB
python:3.13-slim    70729b46c69b        204MB         45.5MB
$ docker system df   → Images 296.5MB
$ du -sk /var/lib/docker/containerd/daemon/io.containerd.content.v1.content io.containerd.snapshotter.v1.overlayfs
70668	io.containerd.content.v1.content
219148	io.containerd.snapshotter.v1.overlayfs
```

72.4 MB + 224.4 MB = 296.8 MB: the Images row already includes the compressed content next to the unpacked snapshots. The README said otherwise and was corrected.
