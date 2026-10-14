# Walkthrough: where did the disk space go?

This walkthrough reproduces the three situations dockdu was built for on a test daemon. It uses `scripts/demo-up.sh`, which creates only objects named `dockdu-demo*` and labelled `dockdu.demo=1`, and `scripts/demo-down.sh`, which removes them by name after checking the label. No prune command is involved.

> Status 2026-10-14: run on two disposable daemons (`docker:28-dind`, Engine 28.5.2; `docker:29-dind`, Engine 29.9.0) with `LOG_MB=100`. The expectations below held on both; raw output is in [../verification/v0.2-validation-2026-10-14.md](../verification/v0.2-validation-2026-10-14.md). `demo-down.sh` removed only the labelled objects and kept unlabelled decoys, including one with a demo name.

Requirements: Docker with BuildKit (`docker buildx`), Compose v2, about 400 MB free.

```
scripts/demo-up.sh          # LOG_MB=200 by default
docker system df
dockdu                      # Linux host as root; on Docker Desktop use the container one-liner
scripts/demo-down.sh
```

## A: container logs df does not show

`dockdu-demo-chatty` writes about 200 MB to stdout with the default `json-file` driver. That driver keeps everything in `/var/lib/docker/containers/<id>/<id>-json.log`, with `max-size` unlimited and `max-file` 1 by default ([json-file docs](https://docs.docker.com/engine/logging/drivers/json-file/)). `docker system df` has no row for those files.

What to expect:
- **`docker system df`:** the Containers row stays tiny. It counts only the containers' writable layers.
- **dockdu `container logs on disk`:** about 200 MiB, labelled "not shown by df".
- **dockdu CONTAINER LOGS:**
  - `dockdu-demo-chatty` shows rotation `unbounded`;
  - `dockdu-demo-bounded` shows `bounded 1m × 3` and stays at or below about 3 MiB across 3 files;
  - the suggestion shows the `log-opts` for `daemon.json` (they only affect new containers).

Check by hand:

```
sudo ls -l /var/lib/docker/containers/$(docker inspect -f '{{.Id}}' dockdu-demo-chatty)/
```

## B: build cache that prune cannot free

`docker buildx build` keeps the build's layers as cache records. The records the built image also uses are marked `Shared`.

What to expect:
- **On Docker Engine ≥ 29 (API ≥ 1.52):** df's Build Cache row includes the shared records. The same bytes are also in the Images row.
- **On Engine ≤ 28:** the CLI leaves shared records out of that row.
- **dockdu:**
  - it reports shared, private and the daemon's reclaimable figure;
  - it prints "prune would NOT free X (shared with images or in use)";
  - in the summary it subtracts the shared bytes once, and only if df's row contains them.
- **`docker buildx du`:** shows a larger reclaimable figure, because it counts every record not in use, including shared ones. Pruning a shared record only drops metadata; the bytes stay with the image.

## C: storage per compose project

Project `dockdu-demo` runs two services on one image and writes 32 MiB into the named volume `dockdu-demo_data`. A volume `dockdu-demo-orphan` is created outside compose.

What to expect in COMPOSE PROJECTS:
- **`dockdu-demo`:** 2 containers, the image once (it belongs to one project only), the volume, and its log bytes.
- **`(none)`:**
  - the chatty and bounded containers and their logs;
  - `alpine:3.22`;
  - the orphan volume, flagged "1 orphan volume(s)" with a single `docker volume rm dockdu-demo-orphan` hint.
- **If the overlap line appears:** `alpine:3.22` and `dockdu-demo/app:1` share the alpine base layer. The note then says IMAGES and TOTAL are not additive by at least that amount.

## Results

Recorded 2026-10-14, `LOG_MB=100`.

| | Engine 28.5.2 (overlay2) | Engine 29.9.0 (containerd store) |
|---|---|---|
| `docker system df` Build Cache row | 96 B (shared records left out) | 147.6 MB (shared records included) |
| `docker buildx du` Reclaimable | 67.11 MB (counts shared records) | 147.6 MB (counts shared records) |
| dockdu build cache: shared / private / reclaimable | 64.0 MiB / 96 B / 96 B | 140.7 MiB / 12.0 KiB / 12.0 KiB |
| dockdu summary formula | df + logs | df − 140.7 MiB shared + logs |
| Logs not shown by df | 109.2 MiB (chatty 106.9 MiB unbounded; bounded 2.3 MiB in 3 files at 1m × 3) | same |
| Compose `dockdu-demo` | 2 containers, image 72.1 MiB, volume 32.0 MiB | 2 containers, image 140.7 MiB, volume 32.0 MiB |
| Image layer overlap reported | 8.1 MiB (alpine base) | 12.7 MiB (alpine base) |

On Engine 29 with the containerd store the demo image is 140.7 MiB for a 64 MiB layer of random data plus alpine: the store keeps the compressed content next to the unpacked snapshot. The Engine 28 figure for the same image is 72.1 MiB.

The README GIF (`docs/demo.gif`) was recorded from the Engine 29 run with `vhs demo.tape` (it now records the Engine 29 demo; see [../verification/demo-2026-10-14.md](../verification/demo-2026-10-14.md)).
