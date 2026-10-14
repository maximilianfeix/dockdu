# Why `docker system df` doesn't explain your full disk, and why prune frees less than it says

*Measured on 2026-10-14 on disposable Docker Engine 28.5.2 and 29.9.0 daemons. Every number below is from [the measurement log](../verification/demo-2026-10-14.md); the commands work on any Linux host where you can read `/var/lib/docker` (as root).*

`docker system df` is precise about what it counts. The trouble is what it leaves out, and that "reclaimable" means two different things depending on your Docker version and image store. This article goes through the four places the gap usually hides, with commands that answer each question by hand.

## 1. Container logs are not in any `df` row

The default log driver, `json-file`, has `max-size` unlimited and `max-file` 1 unless you configure it ([docs](https://docs.docker.com/engine/logging/drivers/json-file/)). The log of a chatty container grows until the disk is full. `docker system df` has four rows (Images, Containers, Local Volumes, Build Cache), and log files are in none of them. Containers' SIZE is the writable layer only. Adding logs to `df` has been requested since 2017 ([moby#32205](https://github.com/moby/moby/issues/32205), [moby#38341](https://github.com/moby/moby/issues/38341)).

**Measured:** one container wrote 500 MiB to stdout.

```text
$ docker system df
TYPE            TOTAL     ACTIVE    SIZE      RECLAIMABLE
Images          2         2         147.7MB   0B (0%)
Containers      4         4         24.58kB   0B (0%)
Local Volumes   2         1         33.55MB   0B (0%)
Build Cache     4         0         147.6MB   12.29kB

$ du -ch /var/lib/docker/containers/*/*-json.log | tail -1
535.1M  total
```

The logs are larger than everything `df` reports combined. `docker system prune` does not touch them: in the same setup it reclaimed `0B`, because the containers were running.

**Check by hand:**

```
sudo du -h /var/lib/docker/containers/*/*-json.log* | sort -h | tail
docker inspect -f '{{.Name}} {{.HostConfig.LogConfig.Type}} {{.HostConfig.LogConfig.Config}}' $(docker ps -aq)
```

**Fix:**
- Set `"log-opts": {"max-size": "10m", "max-file": "3"}` in `/etc/docker/daemon.json`, or use the `local` driver (20 MiB × 5 files, compressed by default).
- Both apply to **new** containers only; existing ones must be recreated.

**Edge cases:**
- **Rotated files** (`-json.log.1`, `.2`) count too.
- **Truncating vs deleting:** deleting a log file of a running container does not free the space until the daemon closes it. A deleted file that is still open shows up in `lsof +L1`, not in `du`. Truncating it to zero frees it at once but discards history.
- **Other drivers:** with `journald` or `syslog` the logs live outside Docker's data root.

## 2. Build cache "shared" with images: counted twice, freed differently

BuildKit marks a cache record `Shared` when its data is also part of an image. Three tools give three different answers for the same records:

| | Engine 28.5.2, overlay2 | Engine 29.9.0, containerd image store |
|---|---|---|
| `docker system df` Build Cache row | leaves shared records out (96 B shown, 64 MiB shared) | includes them (147.6 MB) |
| `docker buildx du` Reclaimable | counts shared records as reclaimable | counts shared records as reclaimable |
| `docker builder prune -af` prints | `Total: 67.11MB` | `Total: 147.6MB` |
| Freed on disk (`du -sxk` before/after) | **≈ 0** | **72.7 MiB of 140.7 MiB** |

Why the rows differ:
- Docker CLI 28 sums only records that are not shared ([`cli/command/system/df.go`](https://github.com/docker/cli/blob/v28.5.1/cli/command/system/df.go), v28.5.1).
- Engine 29 (API ≥ 1.52) sends a total that includes them.

Why `prune` prints a number it does not free:
- On overlay2, a shared record's data belongs to the image, so pruning the record removes metadata.
- On the containerd image store, about half of the shared bytes were freed in two separate runs (72.7 of 140.7 MiB, 56.8 of 108.7 MiB). The build cache evidently kept its own copy of part of the data; these measurements do not show which store objects it consisted of.

What this means in practice:
- **On overlay2,** expect `docker builder prune` to free roughly the *private* part only.
- **On the containerd image store,** expect somewhere between the private part and the printed total. Measure with `du` before and after if it matters.
- **When adding up `df` rows,** Engine 29 counts shared bytes in both the Images and the Build Cache rows.

**Check by hand:**

```
docker buildx du --verbose | grep -E 'ID:|Shared:|Size:'      # per record
docker info --format '{{.Driver}} {{json .DriverStatus}}'     # "io.containerd.snapshotter.v1" = containerd image store
sudo du -sxk /var/lib/docker      # before and after a prune: what it really freed
```

Fresh Engine 29.9.0 installs use the containerd image store by default (observed in `docker:29-dind`), and so does Docker Desktop. Hosts upgraded from older versions may still be on overlay2.

## 3. Image sizes overlap

`docker image ls` shows each image's full size. Two images built `FROM` the same base each include that base, so adding up image sizes, or per-project totals built from them, counts the base more than once. The Images row of `docker system df` counts each layer once.

In the demo the two images share 12.7 MiB. `docker system df -v` shows `SHARED SIZE` and `UNIQUE SIZE` per image; only the unique size comes back when you remove an image that no container uses.

## 4. Docker Desktop: a sparse file

On macOS, all of Docker lives in one VM disk file.
- **Finder and `ls -l` show its maximum size;** `ls -ks` shows what is really allocated ([Docker Mac FAQ](https://docs.docker.com/desktop/troubleshoot-and-support/faqs/macfaqs/)).
- **Freeing space inside the VM** does not shrink the file until Docker Desktop compacts it.
- **Logs are invisible from the host:** the log files from section 1 live inside the VM, where a host-side `du` cannot reach them. Run commands inside the VM instead, for example with `docker run --rm -v /var/lib/docker:/var/lib/docker:ro alpine du …`.

```
ls -ks ~/Library/Containers/com.docker.docker/Data/vms/0/data/Docker.raw
```

## When none of this adds up: is it Docker's at all?

Compare the filesystem with the data root, both in allocated blocks on the same filesystem:

```
df -h /var/lib/docker
sudo du -sxh /var/lib/docker          # -x: do not count overlay mounts of running containers
sudo du -sxh /var/lib/docker/*/ | sort -h
```

What the result tells you:
- **The data root is small:** the full disk is not Docker's.
- **`overlay2/` or `containerd/` is far larger than images + containers + build cache:** look for leftovers from crashes or interrupted pulls ([moby#47089](https://github.com/moby/moby/issues/47089), [moby#52539](https://github.com/moby/moby/issues/52539)).

Don't subtract `docker system df` numbers from `du` numbers: one is logical sizes with shared layers counted once, the other is blocks on disk.

## The one-command version

I wrote [dockdu](https://github.com/maximilianfeix/dockdu) to do the checks above in one read-only run. It:
- measures log files (with `lstat` only);
- splits the build cache into shared and private and says what prune keeps, or that it is unknown on the containerd image store;
- groups everything by compose project;
- with `--disk`, measures the data root next to the filesystem.

Its Docker client refuses any request other than GET/HEAD on four read endpoints, and it never deletes anything.

```
brew install maximilianfeix/tap/dockdu      # or the container image, see the README
```

It does not attribute disk space that no API reports (overlay2 leftovers), it only counts the default BuildKit builder, and it does not run on Windows. The manual commands above work everywhere.

---

*Reproduce every number: `scripts/demo-up.sh` on a disposable daemon, then the commands in [the measurement log](../verification/demo-2026-10-14.md).*
