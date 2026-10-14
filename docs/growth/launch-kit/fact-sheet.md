# Fact sheet

Each line gives its source. Do not state anything that is not here without checking it first.

## The problem

- json-file, Docker's default log driver, has `max-size` unlimited and `max-file` 1 by default ([docs](https://docs.docker.com/engine/logging/drivers/json-file/)).
- `docker system df` has no row for container log files. Upstream requests have been open since 2017 ([moby#32205](https://github.com/moby/moby/issues/32205)) and 2018 ([moby#38341](https://github.com/moby/moby/issues/38341)).
- Build cache records marked `Shared` are also part of images:
  - Engine 29 (API ≥ 1.52) counts them in df's Build Cache row;
  - docker CLI 28 leaves them out (`cli/command/system/df.go`, v28.5.1);
  - `docker buildx du` counts them as reclaimable;
  - **What `docker builder prune -af` actually freed** (data root measured with `du -sxk`):
    - Engine 28.5.2 / overlay2: about 0 of 64 MiB and 28 KiB of 48 MiB;
    - Engine 29.9.0 / containerd image store: 72.7 of 140.7 MiB and 56.8 of 108.7 MiB.

  Observed on 2026-10-14 ([measurement log](../../verification/demo-2026-10-14.md)).
- Docker.raw on macOS: Finder and `ls -l` show the configured maximum, `ls -ks` the real allocation ([Docker Mac FAQ](https://docs.docker.com/desktop/troubleshoot-and-support/faqs/macfaqs/)).
- A fresh Engine 29.9.0 uses the containerd image store by default (observed in `docker:29-dind`, 2026-10-14). The same demo image is 140.7 MiB there and 72.1 MiB on Engine 28.

## What dockdu does

- **Requests:** reads `/_ping`, `/info`, `/system/df` and `/containers/<id>/json`, nothing else. Enforced by a request guard and tests; observed with a logging proxy on both engines.
- **Logs:** measured with `lstat` only, including rotated files; driver and rotation limits are shown.
- **Build cache:**
  - total, shared, private, the daemon's reclaimable figure;
  - "prune would NOT free";
  - shared bytes subtracted once, and only when df's row contains them.
- **Compose:** storage per project. Images used by several projects are counted once; the base-layer overlap between images is reported, not hidden.
- **Volumes:** orphans (unreferenced, not from compose), one `docker volume rm` hint each. dockdu never suggests `volume prune`.
- **Docker Desktop and OrbStack:** apparent vs allocated size of the VM disk file.
- **Output:** `--json` with `schema_version`; `-1` means unknown; exit 2 when anything is partial.
- **Footprint:** two direct Go dependencies (moby client and API types); static binary; 13 MB RSS for a small host run.

## Numbers you may quote

| Claim | Value | Source |
|---|---|---|
| Demo: logs df does not show | 313.6 MiB in df, 537.0 MiB of logs (`du`: 535.1M) | demo-2026-10-14.md §A |
| Prune on shared build cache | overlay2 ≈ 0 freed; containerd 72.7 of 140.7 MiB | demo-2026-10-14.md §B |
| `--disk` vs `du -sxk` | within 0.1% (784.0 MiB vs 802,796 KiB) | demo-2026-10-14.md §A |

| Claim | Value | Source |
|---|---|---|
| Engine 28 bug fixed in v0.2 | v0.1: 149.4 MiB total, 2 false warnings, exit 2; v0.2: 213.4 MiB, exit 0 on the same daemon | validation doc §A |
| Inspect speed-up | 570 → 147 ms for 200 containers at 2 ms per inspect (fake engine) | docs/benchmarks/performance.md |
| Small-host run | 0.01–0.13 s, 13–15 MB RSS | performance.md, validation |

## Do not claim

- That prune never frees build cache shared with images. That holds on overlay2 only; on the containerd image store about half was freed in our runs.

- That dockdu explains *all* disk usage. Overlay2 leftovers and containerd content are not attributed per layer.
- That it is the only tool that shows shared build cache (`buildx du` totals it) or the VM file size (Docker Desktop Settings shows it).
- Support for Windows/WSL2, Colima, OrbStack or rootless as "tested": only context resolution was tested with real CLI-created contexts inside dind.
- Star counts, user numbers or testimonials.
