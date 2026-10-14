# Changelog

All notable changes to dockdu are listed here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Fixed
- Docker Engine 28 and older (API < 1.52): the shared build cache is no longer subtracted twice. Those daemons already leave shared records out of df's build cache row; dockdu now detects which rule the daemon follows (`build_cache.shared_counted_in_df_bytes`, `df_row_excludes_shared`); a row matching neither rule sets `df_row_mismatch` and makes the summary a lower bound. Before, the summary was understated, private cache was wrong, and every run warned and exited 2.
- Compose projects: the header no longer implies that image layers shared between different images are de-duplicated. `projects.image_overlap_bytes` reports by how much the summed image sizes exceed df's image row.
- Volumes whose driver reports no size now make the summary a lower bound (`summary.volume_sizes_unknown`, `summary.complete`).
- Suggested `docker volume rm` commands shell-quote daemon-supplied names, put `--` before a leading dash, and give no command for names with control, bidi-override or line-separator characters.
- A host run on Docker Desktop prints one warning for the invisible data root instead of one per container.
- The fixture recorder creates files with `O_EXCL` and never follows a planted symlink.

### Added
- The Docker client refuses any request that is not GET/HEAD to `/_ping`, `/info`, `/system/df` or `/containers/<id>/json`, before it is sent. Regression tests run against an instrumented fake Engine API.
- dockdu follows the docker CLI's context (DOCKER_CONTEXT, then `currentContext`) when DOCKER_HOST is unset: Colima, OrbStack, rootless Docker, Docker Desktop on Linux. `ssh://` contexts are refused with a hint.
- `schema_version` in `--json` output, with a contract test that fails if a key disappears.
- Container inspects run on four workers (200 containers at 2 ms each: 570 ms → 147 ms on the fake engine).
- `scripts/demo-up.sh` / `scripts/demo-down.sh`: reproducible demo scenarios, removed by name after a label check, never pruned.

## [0.1.0] - 2026-10-13

### Added
- Read-only report of where Docker's disk space went: `docker system df` rows, container log files per container with rotation status, BuildKit cache split into shared and private with the line "prune would NOT free", images with "unique bytes if removed", volumes with referenced / unreferenced / unknown state, everything grouped by compose project.
- `--json` with a stable data contract, `--top`, `--data-root`, `--no-desktop`.
- Docker Desktop and OrbStack section on macOS: apparent vs allocated size of the VM disk file.
- Exit codes: 0 complete, 1 daemon unreachable, 2 report with warnings, 64 usage error.
- Homebrew cask, `go install`, and a `FROM scratch` container image for hosts where the data root is only visible inside the VM.

[Unreleased]: https://github.com/maximilianfeix/dockdu/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/maximilianfeix/dockdu/releases/tag/v0.1.0
