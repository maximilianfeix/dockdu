# Accounting correctness review

Date: 2026-10-14 · Code: branch `v0.2-hardening` · Reviewer: maintainer session

This document reviews every number dockdu prints, what it means, where it comes from, and where it can mislead. Status per finding: **fixed** (with the regression test), **documented** (behaviour kept, wording or docs changed), or **open**.

## 1. What each number is

| Figure | Kind | Source | Precision |
|---|---|---|---|
| `summary.df_*_bytes`, `df_reported_bytes` | Docker-reported logical size | daemon aggregates from `GET /system/df` (API ≥ 1.52) or moby/client's computation from the records (API < 1.52) | as Docker reports it; shared image layers counted once |
| `summary.log_bytes`, `logs.items[].bytes` | measured file bytes | `lstat` of `<root>/containers/<id>/<id>-json.log*` or `local-logs/container.log*` | apparent size (`st_size`), not allocated blocks |
| `summary.docker_bytes_total` | estimate of unique Docker-owned bytes | df − build cache shared with images (only the part df's row contains) + logs | lower bound when `summary.complete` is false |
| `build_cache.reclaimable_bytes` | daemon's reclaimability estimate | daemon (or client for API < 1.52) | what `docker system df` prints; **not** a guarantee of freed disk |
| `build_cache.not_freed_by_prune_bytes` | theoretical bytes a prune keeps | Σ size of records that are shared or in use | exact over the records the daemon returned |
| `images.items[].unique_bytes_if_removed` | theoretical maximum saving for one image | `Size − SharedSize`, only when `Containers == 0` and both known | −1 otherwise, never 0 |
| `projects.*` | attribution | per-image sizes, volume sizes, measured logs | **not additive** for images; see §3.3 |
| `desktop.allocated_bytes` | physical allocation of the VM disk file | `st_blocks × 512` | physical bytes on the macOS volume |

dockdu never claims a number of bytes a prune *will* free on disk. The wording is "reclaimable (as df counts it)" and "prune would NOT free".

Actual reclaimable physical space is not measured, and cannot be from the API alone, because:
- overlay2 and containerd snapshots share data;
- filesystems allocate in blocks;
- the Docker Desktop VM file only shrinks when Desktop compacts it.

## 2. Findings fixed on this branch

### A1 (P0): shared build cache subtracted twice on Docker Engine ≤ 28 — fixed

- **Evidence:**
  - moby/client v0.6.2 `system_disk_usage.go` `buildCacheDiskUsageFromLegacyAPI` computes `TotalSize` as Σ size of records that are **not** shared when the negotiated API is < 1.52.
  - docker CLI v28.5.1 `cli/command/system/df.go` does the same for its Build Cache row (`if !bc.Shared { bsz += bc.Size }`).
  - dockdu v0.1 assumed the row always contains shared records (true on API ≥ 1.52, verified in the recorded 29.2 fixture: total 1,546,143,746 = Σ all records).
- **Impact:** on Engine 28 hosts with shared cache (any host that built images with BuildKit), v0.1 got three things wrong:
  - the summary total was understated by the shared bytes;
  - `private` was understated or shown as unknown;
  - every run carried a "total differs from records sum" warning and exited 2.
- **Fix:** `internal/collect/buildcache.go` detects which rule the daemon total follows:
  1. the total equals the records sum: the row contains shared records;
  2. the total equals the unshared sum: the row leaves them out;
  3. neither: fall back on the API version.

  It reports the records total and exposes `build_cache.shared_counted_in_df_bytes`, the only amount the summary subtracts. The subtraction is clamped to the df row.
- **Tests:**
  - `TestBuildCacheLegacyAPITotalExcludesShared` (API 1.51, 1.44 and unknown);
  - `TestBuildCacheCurrentAPICountsSharedInDf`;
  - `TestBuildCacheSharedSubtractionNeverExceedsDfRow`;
  - `TestBuildLegacyDaemonSummaryDoesNotSubtractSharedTwice`;
  - `TestWriteTextLegacyDaemonSaysNothingWasSubtracted`;
  - `FuzzBuildCache`, 1.4 M executions, no failure.
- **Verified at runtime on 2026-10-14** on Engine 28.5.2 (dind). df's Build Cache row was 96 B with 64 MiB shared. v0.1 printed total 149.4 MiB, private unknown, 2 false warnings and exit 2; the fix prints 213.4 MiB, private 96 B, exit 0 ([validation](../verification/v0.2-validation-2026-10-14.md)).

### A2 (P1): compose totals overstated by layers shared between different images — documented and quantified

- **Evidence:** `Projects` adds each image's `Size`. `Size` includes layers the image shares with other images (`SharedSize`). Three images built from one 160 MB base add that base three times; df counts it once.
- **Impact:**
  - the v0.1 header "shared images counted once" read as a deduplicated total;
  - project IMAGES and TOTAL columns can exceed df's image row.
- **Fix:** without per-layer data there is no exact attribution, so dockdu now says what it knows:
  - the header names what *is* deduplicated (images used by several projects);
  - the new `projects.image_overlap_bytes` is how much the summed image sizes exceed df's image row (a lower bound on the double counting);
  - the text adds a line saying the columns are not additive.
- **Test:** `TestProjectsReportImageLayerOverlap`, golden file updated.

### A3 (P1): unknown volume sizes silently missing from the total — fixed

- **Evidence:** volumes of non-`local` drivers have size −1. df and dockdu's total leave them out, but the summary said nothing.
- **Fix:**
  - `summary.volume_sizes_unknown` and `summary.complete` added;
  - the text total says "lower bound: N volume size(s) unknown".
- **Test:** `TestBuildMarksSummaryIncompleteWhenVolumeSizesUnknown`, golden file.

## 3. Reviewed and correct (no change)

1. **Build cache invariant** `reclaimable + Σ(shared ∨ in use) = Σ records` holds for both the engine rule (moby `builder-next/builder.go`) and the legacy client computation. It is now checked against the records sum, not the df row.
2. **Duplicate IDs:**
   - build-cache records are deduplicated by ID, with a warning;
   - images and volumes are deduplicated by ID or name.

   Containers are not deduplicated; the daemon's list is keyed by ID.
3. **Unknown is never zero:**
   - image unique size, volume size, log bytes and private cache use −1, rendered as "unknown" or "unmeasured";
   - the one remaining exception is an image with `Size < 0` in compose attribution, which adds 0 (open, A6).
4. **Logs:**
   - only regular files are counted, via `lstat`, so symlinks are never followed;
   - rotated siblings `.N` and `.N.gz` are included;
   - a file rotated away between `ReadDir` and `Lstat` is skipped, a sub-second undercount;
   - an unreadable directory makes the item unmeasured and the summary a lower bound.
5. **Docker Desktop:**
   - apparent size (configured maximum) and allocated size are shown as separate numbers;
   - they are never added to or subtracted from Docker's numbers.
6. **Containers removed during the scan:** the inspect fails, and that container's logs are unmeasured with a warning (`TestSnapshotInspectFailureIsPartialNotFatal`).
7. **Integer overflow:** byte sums are `int64`, which holds 8 EiB. Realistic daemons cannot overflow it. A hostile daemon could send sizes near 2⁶³; that would show nonsense numbers but cannot crash dockdu (no panics on arithmetic). The fuzz test bounds inputs to ≤ 2⁵⁰ per record.

## 4. Open accounting questions

| ID | Sev | Finding | Next step |
|---|---|---|---|
| A4 | P1 | Containerd image store (**verified 2026-10-14: default on a fresh Engine 29.9.0**, dind) keeps compressed content plus unpacked snapshots; the demo image shows 140.7 MiB on 29 vs 72.1 MiB on 28. df's image figure may understate disk use, and dockdu inherits that. | Record a fixture on a containerd-store Linux host. Compare `df` with `du -sx /var/lib/containerd /var/lib/docker`. Decide whether to show a "not attributed" line. |
| A5 | P2 | Log `bytes` is `st_size` (apparent), not allocated blocks. For sparse or compressed filesystems these differ. | Add `allocated_bytes` from `st_blocks` on Linux if users report a mismatch. |
| A6 | P3 | An image with `Size < 0` (never observed) counts as 0 in compose attribution. | Count it as unknown like volumes if it is ever observed. |
| A7 | P2 | Overlay2 directories not referenced by any image or container (orphaned by crashes, moby #47089) are invisible to the API. | Out of scope for an API-only tool. Document `du` steps in the walkthrough instead. |

## 5. How to verify

```
go test ./internal/collect -run 'BuildCache|Projects|Build' -v
go test ./internal/collect -run XXX -fuzz FuzzBuildCache -fuzztime 30s
go test ./cmd/dockdu -run JSONContract -v
```

## 6. Found during release validation (2026-10-14)

**A8 (P1): the image "reclaimable (as df counts it)" label was false on Engine 28. Fixed in edd226c.**

- docker CLI 28 computes the image row's reclaimable figure itself: LayersSize − Σ unique bytes of images in use. That counts shared base layers of images that are in use as reclaimable: 8.54 MB in the demo, where the API reports 0 B.
- dockdu keeps the API figure, which is the honest one, and labels it "reclaimable per the Docker API".
- The build cache figure matched df on both engines, so its label stays.
