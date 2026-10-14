## What and why

## Checks

- [ ] `make check` passes (gofmt, vet, tests, build)
- [ ] No new Docker API call, or it is read-only, added to the guard allowlist in `internal/engine/guard.go` and to the fake-engine tests
- [ ] Unknown values stay `-1` / "unknown", never `0`; partial results say "lower bound"
- [ ] JSON: only added keys (`DOCKDU_UPDATE_KEYS=1 go test ./cmd/dockdu -run JSONContract`), or a `schema_version` bump explained here
- [ ] No suggestion that removes volumes in bulk (`volume prune`, `--volumes`)
