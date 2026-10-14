# Product and engineering audit

Date: 2026-10-14 · Base: `main` @ 88ea155 (v0.1.0 + two fixes) · Work branch: `v0.2-hardening` (local, not pushed)

Severity:
- **P0:** security, data correctness or critical failure.
- **P1:** high-impact reliability, installation or UX problem.
- **P2:** valuable improvement.
- **P3:** optional.

Detailed reviews: [accounting](accounting-correctness.md), [security](security-and-readonly-contract.md), [competition](competitive-landscape.md), [performance](../benchmarks/performance.md).

## State found

- **Released:** v0.1.0 on 2026-10-13: GitHub release, ghcr.io image, Homebrew cask (pushed by hand after a token failure).
- **Since the release:** `main` has two fixes (module version for `go install`, release preflight for the tap token).
- **Open PRs:** four Dependabot bumps of `github/codeql-action` v3 → v4 (#18–#21).
- **Repo:** 0 stars, 0 forks, no issues.
- **Tests:** `go build`, `go vet`, `go test` and `gofmt` all clean.
  - Unit tests run from a fixture recorded on Docker Desktop 29.2.
  - One integration test (log rotation) runs in CI inside `golang:1.27`.
- **Claims checked against the code:** all 12 documented capabilities exist. Two descriptions were imprecise, findings P1-5 and P1-6.

## Findings

| ID | Sev | Finding | Evidence | User impact | Resolution | Verification | Status |
|---|---|---|---|---|---|---|---|
| P0-1 | P0 | Shared build cache subtracted twice on Docker Engine ≤ 28 | moby/client v0.6.2 legacy `buildCacheDiskUsageFromLegacyAPI`; docker/cli v28.5.1 `system/df.go` | Wrong totals and private cache, exit 2 on every run on many Linux servers | Detect the daemon's rule; `shared_counted_in_df_bytes` | `TestBuildCacheLegacy*`, `FuzzBuildCache` | **fixed** (d1c97e7) |
| P1-1 | P1 | Read-only promise not enforced, only followed | engine called three endpoints by convention; no test would catch a new mutating call | Trust: the core promise rested on code review | Request-hook guard plus fake-Engine regression tests | `TestGuardRefusesMutatingCalls` | **fixed** (38675d6) |
| P1-2 | P1 | Docker contexts ignored | `client.FromEnv` only; `docker context use colima` left dockdu on `/var/run/docker.sock` | "Cannot reach daemon" for Colima, OrbStack, rootless and Desktop-on-Linux users while `docker ps` works: a first-run failure | `ResolveTarget`: DOCKER_HOST → DOCKER_CONTEXT → currentContext | `TestResolveTarget*`; manual run against the `desktop-linux` context | **fixed** (c3bf437) |
| P1-3 | P1 | Daemon-supplied volume names pasted raw into suggested commands | `SuggestVolumeRm("x; rm -rf ~")` | A copy-pasted suggestion could run a second command | Shell-quote, `--`, refuse names with control characters | `TestSuggestVolumeRmQuotesDaemonNames` | **fixed** (943da34, f375ef5) |
| P1-4 | P1 | Compose image totals double-count shared base layers, and the header implied deduplication | `Projects` sums `image.Size` | Overstated per-project storage | Header wording, `image_overlap_bytes`, "not additive" line | `TestProjectsReportImageLayerOverlap` | **fixed** (02df5b1) |
| P1-5 | P1 | README comparison table: missing `docker buildx du`, lazydocker "compose: no" misleading, Doku cells inconsistent, Docker Desktop's own disk view not acknowledged | [competitive-landscape.md](competitive-landscape.md) | Credibility with expert readers (HN, r/docker) | Rewritten table with verified cells and a row where dockdu says "no" | Review against the research doc | **fixed** (README) |
| P1-6 | P1 | `:ro` socket mount presented without saying it does not restrict the API | README/SECURITY | Users may believe `:ro` makes the socket safe | Explicit warning in README and SECURITY.md | Doc review | **fixed** |
| P1-7 | P1 | Homebrew tap token lacks Contents: write | PROGRESS.md; v0.1.0 release run 403 | Next release fails at the preflight (by design) until fixed | Owner rotates the fine-grained token (needs approval, secret) | Release preflight passes | **blocked: owner action** |
| P1-8 | P1 | Containerd image store: df may understate disk usage and dockdu inherits it | forum t/151053; moby #52539 | The df-vs-disk gap on containerd hosts is partly unexplained | Record a fixture on a containerd-store Linux host and decide on a "not attributed" line. Containerd is now **verified as the default on a fresh Engine 29.9.0** (dind, 2026-10-14) | Fixture plus `du` comparison | **open** |
| P1-9 | P1 | IMAGES header said "reclaimable (as df counts it)", false on Engine 28 (CLI 28 computes its own figure) | Engine 28.5.2 run: df 8.54 MB, API 0 B | Misleading label | Label "per the Docker API" | validation run, golden file | **fixed** (edd226c) |
| P2-1 | P2 | Sequential inspects | `moby.go` loop | Slow on hosts with hundreds of containers | Four bounded workers, order kept | Benchmark 570 → 147 ms | **fixed** |
| P2-2 | P2 | Unknown volume sizes not flagged in the total | `Summarize` | Silent lower bound | `summary.complete`, `volume_sizes_unknown` | test plus golden | **fixed** |
| P2-3 | P2 | Docker Desktop host run printed one warning per container | `logs.go` | A wall of identical warnings on first run | One counted warning pointing to the container image | `TestLogsInvisibleDataRootWarnsOnce` | **fixed** |
| P2-4 | P2 | JSON had no version or contract test | `model.go` | Silent breaking changes for scripts | `schema_version: 1`, key-path golden, no-null and integer-bytes checks | `TestJSONContract` | **fixed** |
| P2-5 | P2 | Fixture writer stat-then-write race | `fixture.go` | Dev-only: could follow a planted symlink | `O_EXCL` | `TestWriteFixtureDoesNotFollowPlantedSymlink` | **fixed** |
| P2-6 | P2 | Releases unsigned, no SBOM | `.goreleaser.yaml` | Supply-chain trust | cosign keyless and syft in GoReleaser (needs release workflow approval) | snapshot plus verify | open |
| P2-7 | P2 | Demo GIF and README sample come from a v0.1 run on one machine | `docs/demo.gif` | Fine, but no reproducible script | `scripts/demo-up.sh` / `demo-down.sh` plus walkthrough, labelled resources and teardown | run on Engine 28.5.2 and 29.9.0 dind | **fixed: run, GIF re-recorded** |
| P2-8 | P2 | Integration test only covers json-file rotation on the CI daemon | `logs_integration_test.go` | Engine 28 and containerd store unverified at runtime | CI matrix with `docker:28-dind` | CI run | open |
| P2-9 | P2 | Dependabot codeql-action v4 PRs open | #18–#21 | Stale actions | Review and merge (owner) | CI green | open |
| P3-1 | P3 | Named buildx builders not measured | SPEC §4 | Users with `docker-container` builders see their cache only as a volume | Explain in the build cache section when a `buildx_buildkit_*` volume exists | fixture test | open |
| P3-2 | P3 | Windows/WSL | SPEC §12 | Not supported | See §"Platform boundaries" | n/a | open |
| P3-3 | P3 | `ReadDir` follows symlinked directory components | security S3 | Low: only sizes of files with a fixed name prefix | `os.Root` | unit test | open |

## Platform boundaries (documented, not changed)

Connecting to a daemon is not the same as seeing its files.

| Setup | API | Log files | Desktop VM file |
|---|---|---|---|
| Linux host, root | ✔ | ✔ | n/a |
| Linux host, user in `docker` group | ✔ | ✘ (0640 root files) → unmeasured, one warning | n/a |
| Rootless Docker (context or DOCKER_HOST) | ✔ (context support, **tested only with synthetic context files**) | ✔ when the data root under `~/.local/share/docker` is readable by the user (**NOT RUN**) | n/a |
| Docker Desktop macOS, host binary | ✔ | ✘ inside VM → one warning | ✔ |
| Docker Desktop macOS, container image | ✔ | ✔ (VM's `/var/lib/docker` mounted read-only) | ✘ (by design) |
| Colima / OrbStack via context | ✔ (**not tested on a real install**) | ✘ inside VM | OrbStack `data.img` ✔; Colima ✘ |
| Remote `tcp://` daemon | ✔ | ✘: local paths are not the daemon's; container IDs will not exist locally, so the result is unmeasured, never wrong numbers | ✘ |
| `ssh://` context | refused with a hint | | |
| Windows / WSL2 | not built | | |

**Windows/WSL feasibility:**
- A `windows/amd64` build of the API part is straightforward. moby/client supports `npipe://`, and contexts with `npipe://` are already accepted.
- Measuring logs from Windows is not possible: the files live inside the WSL2 VM's ext4 disk.
- Inside a WSL2 distro with Docker Engine installed natively, the Linux binary already works.
- For Docker Desktop on Windows, the container one-liner should work as on macOS. **NOT RUN.**
- Recommendation: document the container one-liner for Windows. Build a Windows binary only if users ask for it.

## Product

- **Positioning:** "Where did your Docker disk space go?" plus "dockdu shows what `docker system df` misses, without deleting anything." Both claims hold:
  - logs are not in any df row;
  - shared build cache is in two rows on API ≥ 1.52;
  - nothing is deleted, and that is now enforced.
- **Avoid:**
  - "explains all your disk usage": containerd content, overlay2 orphans and non-Docker files are outside the API;
  - "the only tool showing X": `buildx du` shows shared/private totals and Docker Desktop shows the VM file's real size.
- **Three defensible reasons to choose dockdu** (from the competitive review):
  1. one read-only command that reconciles df with logs and shared cache;
  2. log files measured with rotation and attributed to compose projects;
  3. safe and scriptable: it never deletes, needs no daemon or port, has versioned JSON, and uses exit code 2 for partial data.
- **Time to first value:**
  - Linux root: one command, under a second on small hosts.
  - Docker Desktop: the host binary gives the summary and the VM file immediately; logs need the container one-liner, and the single warning now says so.
