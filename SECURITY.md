# Security

dockdu is read-only by design. It calls three Docker Engine API endpoints (`info`, `system/df`, `containers/{id}/json`) plus the client's version ping, stats log files under the data root, and never removes, stops, prunes or starts anything. If you ever see it do otherwise, that is a bug of the highest severity.

## Reporting a vulnerability

Please report security issues privately through [GitHub's private vulnerability reporting](https://github.com/maximilianfeix/dockdu/security/advisories/new). You will get an answer within a week. Please do not open a public issue for a vulnerability.

## What to know when you run it

- The binary needs the Docker socket. Anything with socket access is effectively root on that host; dockdu uses it only for the three read calls above, but you are trusting the binary. Build from source if that matters to you: `go install github.com/maximilianfeix/dockdu/cmd/dockdu@latest`.
- The container image runs as root inside the container so it can read `root:root 0640` log files. Mount the data root read-only: `-v /var/lib/docker:/var/lib/docker:ro`.
- The macOS section only stats one or two files in your home directory (the Docker Desktop and OrbStack VM disk images).
- No network access, no telemetry, no configuration file.

## Release verification

Releases are built by GitHub Actions from a tag and published with GoReleaser; `checksums.txt` is attached to every release. Verify an archive with `shasum -a 256 -c checksums.txt --ignore-missing`.
