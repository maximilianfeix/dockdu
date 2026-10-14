# Releasing

A release is cut by the maintainer only. Tagging publishes the GitHub release, the ghcr.io image and the Homebrew cask, so nothing in this list is done by automation or contributors without approval.

## Before tagging

1. `main` is green in CI (lint, build, test, integration, CodeQL).
2. Locally: `test -z "$(gofmt -l .)" && go vet ./... && go test -race ./...`.
3. Fuzz smoke test: `go test ./internal/collect -run XXX -fuzz FuzzBuildCache -fuzztime 30s`.
4. JSON contract: `go test ./cmd/dockdu -run JSONContract`. If a key was removed or changed meaning, `report.SchemaVersion` is bumped and the CHANGELOG says so.
5. Dry run: `goreleaser release --snapshot --clean`. Check that `dist/` has 4 archives, `checksums.txt` and two images.
6. Smoke test of the snapshot binary:
   - `./dist/dockdu_darwin_arm64*/dockdu --version`;
   - one real run on a Linux host as root (exit 0);
   - one container run on Docker Desktop.
7. The `HOMEBREW_TAP_GITHUB_TOKEN` secret has Contents: read and write on `maximilianfeix/homebrew-tap`. The release workflow's preflight fails before publishing otherwise.
8. CHANGELOG: move `[Unreleased]` to the new version with the date; update the compare links.
9. README numbers and the demo GIF still match the current output (`vhs demo.tape` if not).

## Tag

```
git tag -a vX.Y.Z -m "dockdu vX.Y.Z" && git push origin vX.Y.Z
```

## After

- Check that the release has 4 archives and `checksums.txt`.
- Run `docker pull ghcr.io/maximilianfeix/dockdu:vX.Y.Z`.
- Run `brew upgrade dockdu` (or `brew install maximilianfeix/tap/dockdu`), then `dockdu --version`.
- Record the verification in `docs/verification/release-<date>.md` and the state in PROGRESS.md.

## v0.2.0 status (2026-10-14, branch `v0.2-hardening`)

Detailed results: [verification/v0.2-validation-2026-10-14.md](verification/v0.2-validation-2026-10-14.md).

| Step | Status |
|---|---|
| 1. CI green on `main` | NOT RUN: the branch is not pushed |
| 2. gofmt, vet, `go test -race` | PASS |
| 3. Fuzz smoke test | PASS |
| 4. JSON contract (only additions; `schema_version` 1) | PASS |
| 5. GoReleaser snapshot | PASS without images (`--skip=docker`); the image build was NOT RUN, but the Dockerfile was built and run in dind (PASS) |
| 6. Smoke test: Linux as root, container one-liner | PASS on Engine 28.5.2 and 29.9.0 (dind); the darwin snapshot binary runs on the host |
| 7. Tap token | BLOCKED: owner action |
| 8. CHANGELOG | PASS ([Unreleased] filled in; move it to 0.2.0 on tagging) |
| 9. README numbers and GIF | PASS (re-recorded from Engine 29.9 dind) |
