# Progress

A fresh session continues from here. Read SPEC.md and CLAUDE.md first.

## Current phase

v0.1.0 released on 2026-10-13 from `ebf8794`. **v0.2 hardening validated** on the local branch `v0.2-hardening` (2026-10-14, **not pushed**, no PR yet). Release verdict: NOT READY to tag until it is pushed, CI is green and the tap token is fixed; the code passed every check that was run. Audit and plan: `docs/audits/product-and-engineering-audit.md`. SPEC.md §13 lists the contract additions.

### v0.2-hardening: done and verified (2026-10-14)

- P0-1: shared build cache no longer subtracted twice on Engine ≤ 28. Source-verified (moby/client v0.6.2 legacy path, docker/cli v28.5.1 `system/df.go`), unit and fuzz tested. **Not run against a real Engine 28 daemon.**
- P1-1: request guard (GET/HEAD on `/_ping`, `/info`, `/system/df`, `/containers/<hex>/json`) plus fake-Engine regression tests for 13 mutating or content-reading calls.
- P1-2: docker context resolution (DOCKER_HOST → DOCKER_CONTEXT → currentContext). Manually verified against this Mac's `desktop-linux` context; Colima, OrbStack and rootless tested only with synthetic context files.
- P1-3: volume names shell-quoted in suggestions. P1-4: compose image overlap reported. P2: unknown volume sizes mark the summary as a lower bound; 4 inspect workers (570 → 147 ms on the fake engine); one warning for an invisible data root; JSON `schema_version` plus contract test; fixture `O_EXCL`.
- Docs: README (headline, Trust, corrected comparison, jq examples, limits), SECURITY, CONTRIBUTING, RELEASING, CHANGELOG [Unreleased], audits (product, accounting, security, competitive), performance, growth plan, content calendar, community research, demo scripts and walkthrough.
- Checks on the branch head: gofmt clean, `go vet` clean, `go test -race ./...` pass, `FuzzBuildCache` 30 s (1.4 M executions) pass, `goreleaser release --snapshot --clean --skip=docker` pass, snapshot binary smoke-run against Docker Desktop 29.2 exit 0.

## Done

- 2026-10-13: Repository created (github.com/maximilianfeix/dockdu), spec and rules committed on `main`.
- 2026-10-13: Toolchain installed: Go 1.27.2, GoReleaser 2.18.3. Docker Desktop (engine 29.2.0, API 1.53) running.
- 2026-10-13: Phase 0 spike answered in `docs/SPIKE.md` (Q1–Q5 plus `local` driver and Desktop VM file).
- 2026-10-13: Phase 1 skeleton: go.mod (`go 1.26`), `cmd/dockdu` with `--help`/`--version` and tests, Makefile (lint fails on unformatted files), .gitignore, MIT LICENSE, README stub, CI workflow. `make check` green locally.
- 2026-10-13: Phase 2: `internal/engine` (Snapshot model, Moby implementation using Info + DiskUsage + ContainerInspect only, Static engine, fixture load/write with no-overwrite, Sanitize), `internal/report/model.go`, `internal/collect/usage.go` (df totals and summary), fixture `internal/collect/testdata/desktop-29.2.json` recorded from this machine via `DOCKDU_RECORD_FIXTURE`. Tests green.
- 2026-10-13: Phase 3: `internal/collect/buildcache.go` (engine reclaimable, shared, private, in-use, not-freed-by-prune, duplicate IDs counted once, total mismatch warned), `images.go` (dangling, unique bytes only when Containers==0 and SharedSize known, sorted by size), `internal/report/suggest.go` (no volume prune, ever). Tests written first; 18 tests green.
- 2026-10-13: Phase 4: `internal/collect/logs.go` (json-file path from inspect or derived, local driver under local-logs/, rotated siblings summed, Lstat only, symlinks ignored, unreadable → unmeasured + incomplete, other drivers "not on disk"), 8 unit tests, integration test SPEC §9.4 creates its own `dockdu.test=1` container via the SDK and removes it by ID. `make integration` runs it inside golang:1.27 with the socket and `/var/lib/docker:ro` mounted: PASS, 3116 bytes in 3 files observed. On the macOS host the test skips (data root inside the VM).
- 2026-10-13: Phase 5: `compose.go` (projects from the compose label, images attributed through container references, images used by several projects reported per project in SharedImageBytes and counted once at section level so project totals stay additive, dangling and unused images in the (none) bucket, stopped containers older than 7 days), `volumes.go` (referenced / unreferenced / unknown; orphan only when unreferenced and not compose-created; one `docker volume rm` hint per orphan, never bulk), `AttachImageProjects`. Tests first; all green.
- 2026-10-13: Phase 6: text and JSON renderers (golden test), report builder, Docker Desktop section, CLI flags `--json --top --data-root --no-desktop`, exit codes 0/1/2/64. Collector review (phases 3–5) and render review applied: IDs validated, paths under the data root, shared images counted once at section level, unmeasured logs as -1 and counted per project, lower-bound markers, control characters neutralised, effective data root reported.

## Verified results (observed at runtime unless noted)

- Read-only bind mount of `/var/lib/docker` inside a container on Docker Desktop exposes `containers/<id>/*-json.log` with rotated siblings and `local-logs/container.log`; buildkit dir present.
- `local` driver leaves `LogPath` empty; file lives at `containers/<id>/local-logs/container.log`.
- Per-container `--log-opt` values appear in `HostConfig.LogConfig.Config`; daemon-level inheritance confirmed from moby source only (NOT RUN at runtime, no daemon log-opts on this machine).
- `go doc` on moby/moby/client v0.6.2: `DiskUsage`, `DiskUsageOptions{Containers,Images,BuildCache,Volumes,Verbose}`, `BuildCacheDiskUsage.Items []build.CacheRecord`, `ContainerInspect(ctx,id,ContainerInspectOptions{Size})`, `Info(ctx,InfoOptions{})`, `New`, `FromEnv`, `IsErrConnectionFailed`.
- Engine `Reclaimable` for build cache excludes shared records; `docker buildx du` does not. Spec 6.4 amended to use the engine value.
- GoReleaser 2.18.3 rejects `install_steps`; `hooks.post.install` string form parses. Spec 10 amended.
- Docker.raw: apparent 245,106,737,152 B, allocated 9,326,448 × 512 B ≈ 4.45 GiB, matches `df` inside the VM.
- With the containerd snapshotter, image bytes are not under `/var/lib/docker`; Spec 6.10 amended to compare allocated file size with API-reported sizes as separate measures.

## Open problems and blockers

- None blocking. Daemon-level log-opts inheritance remains source-verified only.
- Visual identity (banner, mark, palette) is being produced outside the repo and lands with the README in Phase 7.

## Decided spec deviations

See `docs/SPIKE.md`, section "Spec amendments proposed in this PR" (6.1, 6.4, 6.10, 10). All applied in the Phase 0 PR.

## Test resources owned by this project

- Containers with label `dockdu.test=1`: `dockdu-test-logs` (1f1412f2177f), `dockdu-test-default` (871547ac7ec5), `dockdu-test-local` (34b1d5c11f04).
- Demo compose project `dockdu-demo` (containers 02ebd10c9e0b web, 01db9eef1699 worker, 653a0ccb8469 db; volume `dockdu-demo_pgdata`; compose file in the session scratchpad), labelled `dockdu.test=1`.
- Orphan demo volume `dockdu-orphan-demo` (label `dockdu.test=1`).
- Buildx builder `dockdu-builder` (container `buildx_buildkit_dockdu-builder0`, volume `buildx_buildkit_dockdu-builder0_state`).
- Local snapshot images `ghcr.io/maximilianfeix/dockdu:*` (never pushed).
- Image `alpine:3.22` (kept for the integration test).
Remove containers and volumes by ID/name after checking the label; never prune.

## Local branches waiting to be pushed (2026-10-14, second session)

The auto-mode permission classifier refused `git push` and `gh pr create` in this session ("external system writes"), so nothing below is on GitHub yet. All commits are authored by Maxi (repo-local git identity is set).

| Branch | Base | Content | Checks |
|---|---|---|---|
| `v0.2-hardening` | `origin/main` | the v0.2 release candidate, 19 commits (head 302cdd5) | full gate PASS; validated on Engine 28.5.2 and 29.9.0 |
| `ci-codeql-v4` | `origin/main` | all four codeql-action steps on v4.38.3 (24c54180) plus a Dependabot group; supersedes #18–#21 | actionlint, YAML PASS; CI NOT RUN |
| `community-templates` | `origin/main` | issue form for unexplained numbers (asks for redaction), PR checklist | YAML PASS |
| `launch-kit` | `v0.2-hardening` | `docs/growth/launch-kit/`: fact sheet, outlines, directories, release text, FAQ | docs only |
| `diagnose-disk` | `v0.2-hardening` | filesystem line plus `--disk` (data root measured on disk), docs, this handoff | full gate PASS; `--disk` within 0.1% of `du -sxk` on Engine 28/29 dind; 301k files in 1.3–1.5 s, 16–18 MB RSS |

Push and PR commands (run them yourself, or allow `git push` / `gh pr create` for this repo):

```
cd ~/code/dockdu
for b in v0.2-hardening ci-codeql-v4 community-templates launch-kit diagnose-disk; do git push -u origin "$b"; done
gh pr create --base main --head v0.2-hardening --title "v0.2: correct Engine 28 accounting, enforce read-only, docker contexts" --body-file ~/code/dockdu-release/pr-v0.2-body.md
gh pr create --base main --head ci-codeql-v4 --title "ci: codeql-action v4 for every step, grouped Dependabot updates" --body "All four codeql-action steps must use the same version; #18-#21 each bump one and fail. Supersedes #18-#21."
gh pr create --base main --head community-templates --title "chore: issue form for unexplained numbers, PR checklist" --body "Asks for redacted output, never asks to prune; PR checklist carries the read-only and unknown-is-not-zero invariants."
gh pr create --base v0.2-hardening --head launch-kit --title "docs: launch kit for v0.2.0" --body "Fact sheet, outlines for human-written posts, directory criteria, release text, FAQ."
gh pr create --base v0.2-hardening --head diagnose-disk --title "feat: data root filesystem and --disk" --body "Shows the filesystem holding the data root; --disk measures the data root like du -sx and reports what lies outside it. Verified against du on Engine 28/29."
```

Next three actions:
1. Push the branches and open the PRs above, watch CI (`gh pr checks <n> --watch`), fix failures. After #ci-codeql-v4 merges, close #18–#21.
2. Fix the tap token (fine-grained PAT: only `maximilianfeix/homebrew-tap`, Contents read and write), merge v0.2, run `docs/RELEASING.md` for v0.2.0.
3. Retarget `launch-kit` and `diagnose-disk` to `main` after v0.2 merges. Release `--disk` as v0.3.0 once it has been run on a real Linux host with a large data root.

## Test status of the v0.2 branch (release validation 2026-10-14)

Full table and raw outputs: `docs/verification/v0.2-validation-2026-10-14.md`. Environment: two disposable daemons, `docker:28-dind` (Engine 28.5.2, API 1.51) and `docker:29-dind` (Engine 29.9.0, API 1.56, containerd store), removed afterwards.

| Check | Status |
|---|---|
| build, vet, gofmt, `go test -race` (host) | PASS |
| unit tests as Linux binaries in dind | PASS |
| staticcheck 0.8.2 | PASS (after 332330d) |
| govulncheck 1.9.0 | PASS (no vulnerabilities) |
| FuzzBuildCache | PASS |
| cross-compile linux/darwin/windows/freebsd | PASS (compile only) |
| integration test on Engine 28.5.2 and 29.9.0 | PASS |
| P0-1 on a real Engine 28 (v0.1 wrong, branch right) | PASS |
| read-only: real-daemon request log | PASS (only `/_ping`, `/info`, `/system/df`, `/containers/<id>/json`) |
| demo up/down on both engines, decoys survive teardown | PASS |
| container image from the Dockerfile, README one-liner | PASS (Engine 28 dind) |
| real CLI-created contexts (tcp, ssh, missing) | PASS |
| published v0.1.0 archives, checksums, cask sha256, ghcr platforms | PASS |
| `go install @latest` | PASS; known: v0.1.0 prints `dev`, fixed for the next tag |
| GoReleaser snapshot without images | PASS |
| GoReleaser image build, brew install from the tap | NOT RUN |
| tap token preflight | BLOCKED (owner) |
| Colima / OrbStack / rootless real installs, Windows/WSL2, large hosts | NOT RUN |
| CI on the branch | NOT RUN (not pushed) |

Found and fixed during validation: the IMAGES "reclaimable (as df counts it)" label was false on Engine 28 (edd226c).

## Next executable steps (in order)

1. Owner: review the branch and approve a push plus PR. CI must pass on the PR: lint, build, test, integration and CodeQL.
2. Owner: give `HOMEBREW_TAP_GITHUB_TOKEN` Contents: read and write on `maximilianfeix/homebrew-tap`.
3. Merge, then follow `docs/RELEASING.md` for v0.2.0. Move the CHANGELOG [Unreleased] section to 0.2.0; it is the release-notes text.
4. Delete the leftover local branch `hardening-v0.2` after the merge (ask first).
5. After the release: `docs/growth/launch-plan.md`, written by the maintainer.

## Previous next step (still open)

Before the next tagged release: give the fine-grained token behind `HOMEBREW_TAP_GITHUB_TOKEN` Contents: Read and write on `maximilianfeix/homebrew-tap` (the v0.1.0 run failed at the cask step with 403 and the cask was pushed by hand). Optional: upload `docs/social-preview.svg` as PNG under Settings → Social preview.
