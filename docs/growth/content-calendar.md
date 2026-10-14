# Content calendar

Each piece teaches something true and useful without dockdu. dockdu appears as one of the ways to check, next to the manual commands. Write each one yourself; the outlines and verified facts below are the starting point. Do not repost the same text in several places: adapt it or link to it.

| # | Week | Title (working) | Channel | Status |
|---|---|---|---|---|
| 1 | after v0.2 tag | Why `docker system df` says 11 GB when your disk says 138 GB | dev.to (canonical), cross-post Hashnode | outline below |
| 2 | +2 | How unbounded Docker logs fill a server (and the two-line fix) | dev.to; answer links on the Docker forum where relevant | outline below |
| 3 | +4 | Why Docker build cache size is not what prune gives back | dev.to; Golang Weekly / DevOps Weekly pitch | outline below |
| 4 | +6 | What Docker Desktop's Docker.raw size really means | dev.to | outline below |
| 5 | +8 | Attributing Docker storage to compose projects | dev.to, r/selfhosted (rules permitting) | needs real data |

## 1. "docker system df says 11 GB…"

1. **The symptom:** `df -h` shows the disk full while `docker system df` shows a small total.
2. **The four places the gap hides:** log files, build cache counted twice, Docker.raw sparse size, and data the API does not see (overlay2 leftovers, containerd content).
3. **Commands for each by hand:**
   - `sudo du -sh /var/lib/docker/containers/*/*-json.log*`
   - `docker buildx du --verbose`
   - `ls -ks Docker.raw`
   - `sudo du -sx /var/lib/docker /var/lib/containerd`
4. **dockdu** as the one-command version of steps 1–3, with its limits stated.
5. **Proof:** real output from `scripts/demo-up.sh` on a named engine version. **Run it first; no invented numbers.**

## 2. Unbounded logs

- json-file defaults: unlimited size, 1 file.
- `daemon.json` `log-opts` apply to new containers only. Existing containers must be recreated.
- The `local` driver: 20m × 5 files, compressed by default.
- Never truncate the log file of a running container blindly. Explain why: the daemon holds an open file handle.
- Reproducible with `scripts/demo-up.sh` scenario A.

## 3. Build cache vs reclaimable

- What a "Shared" record is.
- The engine rule: reclaimable = !InUse ∧ !Shared.
- buildx du counts !InUse.
- CLI 28 vs 29 row semantics.
- Why pruning shared records frees ~0 bytes.
- Code references:
  - moby `daemon/internal/builder-next/builder.go`;
  - moby/client `system_disk_usage.go`;
  - docker/cli v28.5.1 `cli/command/system/df.go`.

## 4. Docker.raw

- Apparent vs allocated size: `ls -l` vs `ls -ks`.
- Docker Desktop's own Settings view shows the same split. Credit it.
- When the file shrinks: compaction. What lowering the disk limit does: it deletes everything.
- OrbStack `data.img` works the same way.

## 5. Storage per compose project

- Images carry compose labels only when compose built them, so attribution goes through the containers that reference them.
- Base layers shared between images make per-project totals non-additive; say so.
- Needs real before/after data from a user who agrees to share it.
