# Contributing

Thanks for looking. The scope of v0.1 is fixed in [SPEC.md](SPEC.md); anything outside it is welcome as an issue first.

## Ground rules

- dockdu never changes Docker state. A pull request that adds a mutating call will not be merged, however convenient.
- Unknown is unknown. A size that could not be measured is `-1` / "unknown", never `0`.
- Standard library first. A new direct dependency needs a sentence in the pull request explaining why.
- Tests before implementation for every rule in SPEC.md §6.4 to §6.8. Tests are never weakened to make a build green.

## Working on it

```
make check        # gofmt (fails on unformatted files), go vet, go test, build
make integration  # live-daemon test inside golang:1.27 with the socket and /var/lib/docker:ro mounted
make snapshot     # goreleaser release --snapshot --clean
```

Fixtures under `internal/collect/testdata/` are recorded with `DOCKDU_RECORD_FIXTURE=<path> ./dockdu` and sanitised (compose labels only, registry hosts and build-step arguments removed). Recording never overwrites an existing file.

Commits follow Conventional Commits (`feat:`, `fix:`, `test:`, `docs:`, `chore:`). One topic per pull request.
