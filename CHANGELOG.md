# Changelog

All notable changes to dockdu are listed here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added
- The filesystem holding the data root (used, size, free) when the data root is visible, and `--disk`: a `du -sx`-style walk of the data root (lstat only, hard links once, other filesystems not entered, bounded) that reports allocated bytes per top-level entry and the space outside the data root on the same filesystem. JSON: `filesystem`, `filesystem.data_root_on_disk`.

## [0.2.0] - 2026-10-14

### Fixed
- Docker Engine 28 and older (API < 1.52): the shared build cache is no longer subtracted twice. Those daemons already leave shared records out of df's build cache row; dockdu now detects which rule the daemon follows (`build_cache.shared_counted_in_df_bytes`, `df_row_excludes_shared`); a row matching neither rule sets `df_row_mismatch` and makes the summary a lower bound. Before, the summary was understated, private cache was wrong, and every run warned and exited 2.
- Compose projects: the header no longer implies that image layers shared between different images are de-duplicated. `projects.image_overlap_bytes` reports by how much the summed image sizes exceed df's image row.
- Volumes whose driver reports no size now make the summary a lower bound (`summary.volume_sizes_unknown`, `summary.complete`).
- Suggested `docker volume rm` commands shell-quote daemon-supplied names, put `--` before a leading dash, and give no command for names with control, bidi-override or line-separator characters.
- A host run on Docker Desktop prints one warning for the invisible data root instead of one per container.
- The fixture recorder creates files with `O_EXCL` and never follows a planted symlink.
- The IMAGES header says "reclaimable per the Docker API" instead of "as df counts it": docker CLI 28 prints its own, different image figure.
- `go install github.com/maximilianfeix/dockdu/cmd/dockdu@vX.Y.Z` builds report their module version instead of `dev` (already on `main` after 0.1.0).

### Added
- The Docker client refuses any request that is not GET/HEAD to `/_ping`, `/info`, `/system/df` or `/containers/<id>/json`, before it is sent. Regression tests run against an instrumented fake Engine API.
- dockdu follows the docker CLI's context (DOCKER_CONTEXT, then `currentContext`) when DOCKER_HOST is unset: Colima, OrbStack, rootless Docker, Docker Desktop on Linux. `ssh://` contexts are refused with a hint.
- `schema_version` in `--json` output, with a contract test that fails if a key disappears.
- Container inspects run on four workers (200 containers at 2 ms each: 570 ms → 147 ms on the fake engine).
- `scripts/demo-up.sh` / `scripts/demo-down.sh`: reproducible demo scenarios, removed by name after a label check, never pruned.

### Verified
- On disposable Docker Engine 28.5.2 and 29.9.0 daemons: integration test, demo scenarios, the Engine 28 build cache fix, the request set against real daemons (only `/_ping`, `/info`, `/system/df`, `/containers/<id>/json`), and the container image one-liner. Details: `docs/verification/v0.2-validation-2026-10-14.md`.

## [0.1.0] - 2026-10-13

### Added
- Read-only report of where Docker's disk space went: `docker system df` rows, container log files per container with rotation status, BuildKit cache split into shared and private with the line "prune would NOT free", images with "unique bytes if removed", volumes with referenced / unreferenced / unknown state, everything grouped by compose project.
- `--json` with a stable data contract, `--top`, `--data-root`, `--no-desktop`.
- Docker Desktop and OrbStack section on macOS: apparent vs allocated size of the VM disk file.
- Exit codes: 0 complete, 1 daemon unreachable, 2 report with warnings, 64 usage error.
- Homebrew cask, `go install`, and a `FROM scratch` container image for hosts where the data root is only visible inside the VM.

[Unreleased]: https://github.com/maximilianfeix/dockdu/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/maximilianfeix/dockdu/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/maximilianfeix/dockdu/releases/tag/v0.1.0
