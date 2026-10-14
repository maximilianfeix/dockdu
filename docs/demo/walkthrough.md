# Walkthrough: where did the disk space go?

This walkthrough reproduces the three situations dockdu was built for on a test daemon. It uses `scripts/demo-up.sh`, which creates only objects named `dockdu-demo*` and labelled `dockdu.demo=1`, and `scripts/demo-down.sh`, which removes them by name after checking the label. No prune command is involved.

> Status 2026-10-14: the scripts are written and syntax-checked but have **not been run**. The "What to expect" notes come from the code and the Docker docs, not from a recorded run. Record real output before using any number from this page publicly.

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

_Not recorded yet. Paste the real `docker system df` and dockdu output here after a run, with the engine version and the date, and regenerate `docs/demo.gif` with `vhs demo.tape`._
