# Security

dockdu is read-only by design and by enforcement. It calls three Docker Engine API endpoints (`info`, `system/df`, `containers/{id}/json`) plus the client's version ping, stats log files under the data root, and never removes, stops, prunes or starts anything. Every Docker client it creates refuses, before sending, any request that is not a GET/HEAD to one of those paths (`internal/engine/guard.go`, tested against a fake Engine API in `internal/engine/fakeengine_test.go`). If you ever see it do otherwise, that is a bug of the highest severity.

Full review: [docs/audits/security-and-readonly-contract.md](docs/audits/security-and-readonly-contract.md).

## Reporting a vulnerability

Please report security issues privately through [GitHub's private vulnerability reporting](https://github.com/maximilianfeix/dockdu/security/advisories/new). You will get an answer within a week. Please do not open a public issue for a vulnerability.

## What to know when you run it

- The binary needs the Docker socket. Anything with socket access is effectively root on that host; dockdu uses it only for the three read calls above, but you are trusting the binary. Build from source if that matters to you: `go install github.com/maximilianfeix/dockdu/cmd/dockdu@latest`.
- Mounting the socket with `:ro` does **not** make the Docker API read-only. It only prevents replacing the socket file; POST and DELETE requests still work over a read-only bind mount. The read-only guarantee is dockdu's own request guard, not the mount option.
- The container image runs as root inside the container so it can read `root:root 0640` log files. Mount the data root read-only: `-v /var/lib/docker:/var/lib/docker:ro`.
- The macOS section only stats one or two files in your home directory (the Docker Desktop and OrbStack VM disk images).
- No network access, no telemetry, no configuration file.

## Release verification

Releases are built by GitHub Actions from a tag and published with GoReleaser; `checksums.txt` is attached to every release. Verify an archive with `shasum -a 256 -c checksums.txt --ignore-missing`.
