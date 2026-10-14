# dockdu v0.3.0

## Fixed: build cache shared with images on the containerd image store

v0.2 printed "prune would NOT free X" for build cache shared with images. That holds on overlay2 but not on the containerd image store, which is the default on fresh Docker Engine 29 installs and on Docker Desktop. Measured on disposable daemons with `docker builder prune -af`:

| Engine / image store | Shared build cache | Freed on disk |
|---|---|---|
| 28.5.2 / overlay2 | 64.0 MiB | ≈ 0 |
| 29.9.0 / containerd | 140.7 MiB | 72.7 MiB |

On the containerd store, dockdu now says that a prune frees an unknown part of the shared bytes and marks the summary as a lower bound. JSON: `build_cache.not_freed_by_prune_bytes` is -1 there, `build_cache.shared_may_be_freed_by_prune` is true, and `daemon.image_store` says which store is in use.

## New: is the full disk Docker's at all?

- When the data root is visible, dockdu prints the filesystem it lives on: used, size, free.
- `dockdu --disk` walks the data root like `du -sx`, with metadata only. It does not follow symlinks, counts hard links once and does not enter running containers' mounts. It shows what each top-level directory occupies and how much of the same filesystem lies outside the data root.
- It is within 0.1% of `du -sxk` on the test daemons.

Details: [measurement log](https://github.com/maximilianfeix/dockdu/blob/main/docs/verification/demo-2026-10-14.md) · [CHANGELOG](https://github.com/maximilianfeix/dockdu/blob/v0.3.0/CHANGELOG.md)

## Install

```
brew upgrade dockdu    # or: brew install maximilianfeix/tap/dockdu
go install github.com/maximilianfeix/dockdu/cmd/dockdu@v0.3.0
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro -v /var/lib/docker:/var/lib/docker:ro ghcr.io/maximilianfeix/dockdu:v0.3.0
```
