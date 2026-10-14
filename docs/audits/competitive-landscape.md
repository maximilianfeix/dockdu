> Research snapshot from 2026-10-14. Star counts, versions and rules change; re-check before quoting any of them publicly.

# dockdu: competitive landscape

Researched 2026-10-14. Each claim is marked **verified** (with the source I read on that date) or **unverified/assumption**. When I say a tool "does not" do something, I checked its current docs or source. Where I could not check, the cell says so.

dockdu reference point: repo `maximilianfeix/dockdu`, release v0.1.0, created 2026-10-13, 0 stars on 2026-10-14. I only read the README and SPEC.

## Versions and commits inspected

| Tool | What I looked at | Stars (2026-10-14) | Last release |
|---|---|---|---|
| `docker system df` | docs.docker.com/reference/cli/docker/system/df/ ; docker/cli `master` @ `c11295a` (2026-10-14), `cli/command/formatter/{disk_usage,buildcache}.go` | docker/cli 6,089 | moby `docker-v29.9.0`, 2026-10-08 |
| `docker buildx du` | docs.docker.com/reference/cli/docker/buildx/du/ | docker/buildx 4,532 | v0.38.0, 2026-10-07 |
| lazydocker | jesseduffield/lazydocker shallow clone @ `7e7aadc` (2026-04-19) | 53,120 | v0.25.2, 2026-04-19 |
| dive | wagoodman/dive shallow clone @ `d6c6919` (2025-05-16) | 54,636 | v0.13.1, 2025-03-29 |
| Doku | amerkurev/doku shallow clone @ `91fb2fc` (2025-12-25), README and `app/` source | 451 | v1.3.0, 2025-12-26 |
| Docker Desktop Disk Usage extension | hub.docker.com/extensions/docker/disk-usage-extension | n/a (closed listing; GitHub repo `docker/disk-usage-extension` returns 404) | 0.2.9, "Updated over 2 years ago", 1M+ installs |
| Docker Desktop settings / Mac FAQ | docs.docker.com/desktop/settings-and-maintenance/settings/ ; docs.docker.com/desktop/troubleshoot-and-support/faqs/macfaqs/ | n/a | n/a |
| Portainer CE | docs.portainer.io/user/docker/volumes and /stacks | 38,662 | 2.45.2, 2026-10-08 |
| ncdu on `/var/lib/docker` | dev.yorhel.nl/ncdu and /ncdu/man | n/a (not on GitHub) | 2.9.2, 2025-10-24 (Zig); 1.22 LTS, 2025-03-05 |
| Others found via `gh search repos "docker disk usage"` | tyutyutyu/dodu (1 star, Go, "ncdu-style TUI/CLI for Docker disk usage", last push 2026-05-01); oullin/docktui (0 stars, 2026-06); Simrayz/docker-analysis (0, 2026-04); AN3S-CREATE/Docker-Space-Analyzer-Reclaimer-CLI (0, prunes) | 0–1 | none notable |

I found no maintained tool called "dutop" or "docker-disk-usage" with any traction. BiffBangPow/docker-disk-usage has 0 stars and no activity since 2018. The real field is the Docker CLI, Doku, Docker Desktop's own UI and extension, and general-purpose managers (Portainer, lazydocker).

## Per-tool findings

### `docker system df` / `docker system df -v`
- **Scope:** images, containers, local volumes, and build cache totals with TOTAL/ACTIVE/SIZE/RECLAIMABLE. *Verified: docs page and `disk_usage.go`.*
- **`-v` adds:** per image `SIZE SHARED SIZE UNIQUE SIZE CONTAINERS`, per container `LOCAL VOLUMES SIZE`, per volume `NAME LINKS SIZE`, and per build cache record `CACHE ID, CACHE TYPE, SIZE, CREATED, LAST USED, USAGE, SHARED`. *Verified: `buildcache.go` line 17 `defaultBuildCacheTableFormat` includes `{{.Shared}}`. The docs example does not show a build cache section, but the source does.*
- **Container log size:** not shown. The docs do not mention logs, and moby issues [#32205](https://github.com/moby/moby/issues/32205) (2017, open) and [#38341](https://github.com/moby/moby/issues/38341) ("Include log space usage in `system df`", 2018, open) confirm this. *Verified.*
- **Build cache shared vs private totals:** a per-record SHARED boolean only. There is no Shared/Private total line. *Verified (source).*
- **Compose attribution:** none. **Orphan volumes:** `LINKS` column, where 0 means unreferenced, with no compose awareness. **Docker Desktop VM file:** no. **Deletes:** no (prune is a separate command). **Output:** table, Go template, and `--format json`. *Verified (docs).*
- **Install friction:** none, it ships with Docker.

### `docker buildx du` (incl. `--verbose`)
- Disk usage of the selected builder, with columns `ID RECLAIMABLE SIZE LAST ACCESSED`. **It prints totals: `Shared`, `Private`, `Reclaimable`, `Total`.** `--verbose` (shorthand for `--format=pretty`) prints per record Parents, Created at, Mutable, Reclaimable, Shared, Size, Description, Usage count, Last used, and Type. It also has `--format json` (NDJSON), `--filter`, and `--builder`. *Verified: docs page.*
- It works with any builder, including named/docker-container builders, which dockdu excludes in v0.1.
- It has no logs, compose, volumes, or Desktop VM information. It does not delete anything (prune is a separate command).
- **Consequence for dockdu:** the shared/private split is not unique to dockdu. The parts that are unique are (a) putting the shared figure next to `system df` and (b) wording it as "prune would NOT free", plus the de-duplicated total.

### lazydocker (v0.25.2)
- TUI for containers, images, volumes, networks, and compose services. *Verified: README and `pkg/gui`.*
- **Log file size:** not shown. It streams log content, but the source has no LogPath/size code (grep for `LogPath`, `SizeRw` found nothing). **Build cache:** no code references (grep `BuildCache` found nothing). *Verified (source grep).*
- **Compose:** it has a `project_panel.go` and `services_panel.go` and reads `com.docker.compose.project` (`pkg/commands/docker.go:420`). So it *is* compose-aware for containers/services in the current directory's project, but it does **not** attribute disk usage per project. *Verified.*
- **Volumes:** shows a volume's `UsageData.Size` in the detail panel. It has no orphan flag, but offers "prune volumes". **Deletes:** yes, on request (prune containers/images/volumes/networks, remove image with confirmation). *Verified (source).*
- **Output:** TUI only, no JSON. **Install:** brew/go/binary/container. **Desktop VM file:** no (no reference in source).

### dive (v0.13.1)
- Explores the layers of a single image and scores image efficiency/wasted space. It has a CI mode (`--ci`, `lowestEfficiency`, `highestWastedBytes`) and **`--json <file>` export** ("Skip the interactive TUI and write the layer analysis statistics", `cmd/dive/cli/internal/options/export.go:27`). *Verified.*
- It has no logs, build cache, compose, volumes, or Desktop VM information, and no delete code (grep for `ImageRemove`/`Prune` found nothing). *Verified.*
- It does a different job (image-size optimization). It is an adjacent tool, not a competitor.

### Doku (v1.3.0) — the closest competitor
- Web UI (Python/uvicorn) that runs as a long-lived container with `-v /var/run/docker.sock:ro -v /:/hostroot:ro -p 9090:9090`. Optional basic auth and SSL. Scans periodically: df every 60 s, logs every 300 s, bind mounts every hour, overlay2 every day. *Verified: README.*
- **Covers more storage than dockdu:** images, containers, volumes, builder cache, **overlay2 storage** (in use vs not in use), **container logs**, and **bind mounts**. *Verified: README "Features", `app/scan/scanner.py`.*
- **Log size per container:** yes, but only the current `LogPath` file (`path.stat().st_size`, scanner.py ~L203). It does **not** include rotated siblings, the driver, or the rotation config. *Verified (source).*
- **Build cache:** a per-record "Shared" yes/no column and an overall total (`templates/pages/build_cache.html`, `context.py build_cache()`). There is no shared/private total. *Verified.*
- **Images:** a Shared Size column. **Volumes:** size plus a list of containers using each volume (`pages/volumes.html`). An unreferenced volume shows an empty list. It has no orphan label and is not compose-aware. *Verified.*
- **Compose attribution:** none. Grep for `compose` and `label` in `app/` found no compose logic. *Verified.*
- **Docker Desktop VM file:** none (no `Docker.raw`/desktop references). *Verified (grep).* Whether its `/:/hostroot` mount measures logs inside the Desktop VM is *unverified*.
- **Deletes:** no. The README says it is "strictly a monitoring tool and never performs any cleanup". *Verified.* **JSON/API:** no documented API. *Unverified whether any JSON endpoint exists.*
- It mounts the whole host root read-only and exposes a port, which is a larger attack surface than a one-shot CLI. That is a factual difference.

### Docker Desktop built-in features and the Disk Usage extension
- **Settings → Resources → Advanced** "displays the maximum size of the disk image and the actual space the disk image is consuming". The Mac FAQ explains that "many tools report the maximum file size, not the actual file size" and suggests `ls -klsh Docker.raw`. *Verified: macfaqs page.* **This means Docker Desktop itself already shows apparent vs actual Docker.raw size.** dockdu adds the same figure to a terminal/JSON report next to the resource totals, and also handles OrbStack's `data.img`.
- **Disk Usage extension** (`docker/disk-usage-extension`, by Docker Inc.): categorizes disk space by images, volumes and similar resources, shows reclaimable space, and offers **one-click reclaim**. It removes stopped containers, unused images, dangling images, build cache, and unused volumes. Version 0.2.9, "updated over 2 years ago", 1M+ installs. *Verified: Hub listing.* Log sizes, compose grouping, and the shared/private build cache split are *not mentioned on the listing (unverified, the source is not public)*.
- I found no release note (through 4.86.0, 2026-08-10) adding a new built-in "Disk usage" dashboard view. *Unverified: absence based on search plus release notes, not exhaustive.*

### Portainer CE (2.45.2)
- Web management UI. The volumes list labels a volume **"unused"** when "Portainer cannot see any applications that are using this volume". It can add, remove, and browse volumes. *Verified: docs.portainer.io/user/docker/volumes.* It has a Stacks concept for compose, but the docs do not mention per-stack disk usage. *Verified that the docs don't say. Whether it exists is unverified.*
- I found no official docs for log file size or for a build cache shared/private split. *Unverified (not found).* **Deletes:** yes. **Install friction:** a long-running server container plus an admin account.

### ncdu on `/var/lib/docker`
- A general disk-usage TUI. Run as root on `/var/lib/docker`, it *will* show `containers/<id>/<id>-json.log*` sizes and overlay2 directories, but with no Docker semantics: no container names, compose, shared vs private, or reclaimability. Overlay2 directory sizes double-count merged mounts if you scan while containers are running (*assumption, from common knowledge*).
- **Can delete:** `d` key "Delete the selected file or directory". `-r` disables deletion. *Verified: ncdu man page.* Deleting under `/var/lib/docker` corrupts Docker state, so this is a real safety difference. JSON export exists (`-o`). It does not work on macOS Docker Desktop without getting a shell inside the VM.

## Comparison table

Legend: V = verified (source in the sections above), U = unverified/assumption.

| | system df -v | buildx du | lazydocker | dive | Doku | DD ext./settings | Portainer | ncdu | dockdu (per its README) |
|---|---|---|---|---|---|---|---|---|---|
| Per-container log size | no (V) | no (V) | no (V) | no (V) | **yes, current file only** (V) | not listed (U) | not found (U) | raw files, no names (V) | yes, incl. rotated |
| Log rotation config shown | no (V) | no (V) | no (V) | no (V) | no (V) | U | U | no (V) | yes |
| BuildKit shared vs private | per-record SHARED flag (V) | **Shared + Private totals** (V) | no (V) | no (V) | per-record flag (V) | reclaimable only (V listing) | U | no | totals + "prune would not free" |
| Compose attribution of storage | no (V) | no (V) | compose-aware UI, no storage per project (V) | no (V) | no (V) | U | stacks exist, sizes U | no | yes |
| Orphan volumes | LINKS=0 (V) | n/a | no flag (V) | n/a | empty container list (V) | "unused volumes" reclaimable (V) | "unused" label (V) | no | yes, compose-aware |
| Docker Desktop VM file | no (V) | no (V) | no (V) | no (V) | no (V) | **max vs actual in Settings** (V) | no (U) | only if pointed at it (assumption) | yes, + OrbStack |
| Overlay2 / bind mounts | no | no | no | per-image layers | **yes** (V) | U | U | raw yes | no |
| Can delete | no | no | yes (V) | no (V) | no (V) | **yes, one-click** (V) | yes (V) | yes, unless `-r` (V) | never |
| Output | table/template/**JSON** (V) | table/pretty/**JSON** (V) | TUI (V) | TUI + **JSON export** (V) | web UI (V) | GUI | web UI | TUI + JSON (V) | text + JSON |
| Install friction | none | none | low | low | daemon container, `/` mount, port | Desktop only | server + account | apt/brew + root | low |
| Stars | 6,089 (cli) | 4,532 | 53,120 | 54,636 | 451 | 1M+ installs | 38,662 | n/a | 0 |

## README "Compared" table: cells that are wrong, misleading, or unverifiable

1. **Missing column: `docker buildx du`.** It is the obvious objection to "Build cache shared vs private: yes". buildx du prints `Shared` and `Private` totals (docs). The dockdu cell is fine, but the table implies nobody else totals them. Add a buildx du column, or reword the dockdu cell to "totals next to df, plus 'prune would not free'".
2. **`docker system df -v` / Form = "CLI"**: incomplete. df supports `--format json` (docs). The dockdu cell "one read-only command, JSON" reads as if JSON is unique to dockdu.
3. **dive / Form = "TUI"**: incomplete. dive has `--json <file>` export and a `--ci` mode (source).
4. **lazydocker / Grouped by compose project = "no"**: misleading. lazydocker reads `com.docker.compose.project` and has Project/Services panels. Suggested: "UI only, no sizes".
5. **Doku / Build cache shared vs private = "partly"**: inconsistent with the df cell. Doku shows exactly what df -v shows (a per-record Shared yes/no plus a total). Use the same wording as df: "per record, no total".
6. **Doku / Orphan volumes = "partly"**: defensible but vague. It lists the containers per volume, and an empty list means unreferenced. It has no label and no compose awareness. Suggested: "lists users per volume, no flag".
7. **Doku / Logs per container = "yes"**: true, but undersells the difference. Doku measures only the current `LogPath` file, with no rotated files, driver, or rotation status. If you want to claim an edge, put it there.
8. **Doku / Form = "web UI in a container"**: correct. It is also a long-running daemon that mounts `/` read-only and opens port 9090, which is worth stating.
9. **"Docker Desktop VM file: no" for every column**: correct for the five tools listed. But Docker Desktop's own Settings → Advanced shows maximum vs actual disk image size (Mac FAQ). The prose "dockdu shows both numbers" is fine; avoid implying no one else does.
10. **`docker system df -v` / Orphan volumes = "unreferenced only, not compose-aware"**: accurate (LINKS column). Optionally say "LINKS count".
11. **lazydocker / Deletes = "on request"**: accurate. **dive / Deletes = "no"**: accurate. **Doku / Deletes = "no"**: accurate.
12. **Missing rows that favour competitors**: overlay2 and bind mount sizes (Doku has them, dockdu does not). An honest table should include at least one row where dockdu says "no".
13. **Missing competitor**: Docker Desktop's Disk Usage extension (1M+ installs, by Docker), which is what Mac users are pointed to first. It deletes with one click, which strengthens dockdu's "never deletes" positioning.
14. README "Why" section: moby #47089 is "open since 2024" (verified: opened 2024-01-17, open, 16 comments). #52539 matches its title, "shows 265.5GB used for images without any images" (opened 2026-05-05, open, 35 comments). Both verified.

## Notes and risks found during research
- **json-file defaults:** `max-size` "Defaults to -1 (unlimited)" and `max-file` "Defaults to 1" (docs.docker.com/engine/logging/drivers/json-file/). This supports the README's "no size limit and no rotation". *Verified.*
- **containerd image store:** a Docker forum moderator (forums.docker.com/t/151053, 2026-02) notes that the containerd image store keeps a compressed and an uncompressed copy of each image, and that df/Portainer figures can understate actual usage. If Engine 29 hosts use the containerd store by default, the "df vs disk" gap there has a cause dockdu does not attribute. *Unverified whether that is the default for new installs. Worth testing before the launch post.*

## Three defensible reasons to choose dockdu
1. **One read-only command that reconciles df with reality.** No other tool I checked computes `df − shared build cache + logs` and states what prune would not free. buildx du has the shared/private totals, but not next to df and not de-duplicated against image sizes.
2. **Log files measured properly and attributed to compose services.** That means rotated files, driver, and bounded/unbounded rotation, plus per-compose-project totals. Doku shows only the current log file and has no compose attribution. lazydocker, dive, df, and buildx du show no log sizes at all.
3. **Safety and scriptability for CI and support threads.** It never deletes (unlike lazydocker, Portainer, ncdu, and the Docker Disk Usage extension), needs no long-running container, mounted host root, or open port (unlike Doku), and has stable JSON with `-1` = unknown plus exit code 2 for incomplete data. That makes it pasteable into forum answers and usable in cron/CI.

## Where competitors are better
- **Doku:** overlay2 in-use vs not-in-use and bind mount sizes, which can be the real culprit when df looks small. It also offers continuous monitoring with a history of scans and a GUI. It has 451 stars and is already listed in awesome-docker.
- **`docker buildx du`:** works with every builder, including named docker-container/remote builders that dockdu v0.1 ignores, and has rich per-record detail with `--filter`.
- **Docker Desktop (Settings and Disk Usage extension):** zero install for Mac users, shows actual vs max disk image size, and reclaims with one click.
- **lazydocker / Portainer:** fix the problem in the same place they show it (prune and remove), and do far more than disk.
- **ncdu:** finds anything on the filesystem, including non-Docker files that share the disk, which dockdu cannot see.
- **Platform reach:** dockdu v0.1 has no Windows/WSL support. Docker Desktop's tools and Portainer work there.
