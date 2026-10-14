# Contributing

Thanks for looking. The v0.1 contract is fixed in [SPEC.md](SPEC.md); changes since then are listed in its §13. Anything larger than a fix is welcome as an issue first.

## Ground rules

- dockdu never changes Docker state. A pull request that adds a mutating call will not be merged. It would also fail: the client's request guard refuses it, and `TestGuardRefusesMutatingCalls` catches it.
- Unknown is unknown. A size that could not be measured is `-1` / "unknown", never `0`, and a total that misses something says "lower bound".
- Standard library first. A new direct dependency needs a sentence in the pull request explaining why.
- Tests come before the implementation for every accounting rule (SPEC.md §6.4–§6.8). Tests are never weakened to make a build green.

## How it fits together

```
cmd/dockdu/main.go         flags, exit codes, picks the daemon (DOCKER_HOST → DOCKER_CONTEXT → current context)
internal/engine/           the only code that talks to Docker
  moby.go                  Info + DiskUsage + ContainerInspect (4 workers) → Snapshot
  guard.go                 request hook: GET/HEAD on four paths, everything else refused
  context.go               reads the docker CLI context metadata
  fixture.go               Snapshot ⇄ JSON fixture, sanitised
internal/collect/          pure functions Snapshot → report sections (only logs.go touches the filesystem, with lstat)
internal/report/           the Report model (what --json prints), text renderer, suggestion strings
```

Data flows one way: `engine.Snapshot` → `collect.Build` → `report.Report` → text or JSON. Collectors never print, renderers never compute.

## Your first change, without a Docker daemon

Everything except the engine is tested from recorded snapshots, so you do not need Docker running:

```
go test ./...                                                  # all unit tests, no daemon needed
go test ./internal/collect -run BuildCache -v                  # one area
go test ./internal/report -run Golden -update                  # after an intended text change; review the diff
DOCKDU_UPDATE_KEYS=1 go test ./cmd/dockdu -run JSONContract    # after adding a JSON field (never to remove one)
```

To test a new rule, build a small `engine.Snapshot` in the test (see `legacyCache` in `internal/collect/buildcache_test.go`) or load the recorded fixture `internal/collect/testdata/desktop-29.2.json`.

## Reproducing a bug report safely

Ask the reporter for `dockdu --json` (it contains no file contents and no environment variables), or for a sanitised snapshot recorded on their machine:

```
DOCKDU_RECORD_FIXTURE=/tmp/report-fixture.json dockdu
```

The recording keeps compose labels only, replaces registry hosts with `<registry>`, strips build-step arguments, and never overwrites an existing file. Commit it under `internal/collect/testdata/` with a test that fails before the fix. Never ask anyone to run a prune command to reproduce a bug.

## Checks

```
make check        # gofmt (fails on unformatted files), go vet, go test, build
make integration  # live-daemon test inside golang:1.27 with the socket and /var/lib/docker:ro mounted
make snapshot     # goreleaser release --snapshot --clean
```

Integration tests create their own containers labelled `dockdu.test=1` and remove them by ID. Run them only against a test daemon.

Commits follow Conventional Commits (`feat:`, `fix:`, `test:`, `docs:`, `chore:`). One topic per pull request. Releases follow [docs/RELEASING.md](docs/RELEASING.md).
