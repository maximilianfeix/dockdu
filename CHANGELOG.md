# Changelog

All notable changes to dockdu are listed here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added
- Read-only report of where Docker's disk space went: `docker system df` rows, container log files per container with rotation status, BuildKit cache split into shared and private with the line "prune would NOT free", images with "unique bytes if removed", volumes with referenced / unreferenced / unknown state, everything grouped by compose project.
- `--json` with a stable data contract, `--top`, `--data-root`, `--no-desktop`.
- Docker Desktop and OrbStack section on macOS: apparent vs allocated size of the VM disk file.
- Exit codes: 0 complete, 1 daemon unreachable, 2 report with warnings, 64 usage error.
- Homebrew cask, `go install`, and a `FROM scratch` container image for hosts where the data root is only visible inside the VM.
