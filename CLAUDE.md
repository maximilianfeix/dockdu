# dockdu

Read-only Go CLI that explains where Docker's disk space went. @SPEC.md is the authoritative scope for v0.1; when this file and SPEC.md disagree, SPEC.md wins.

## Commands

- Build: `go build ./...`
- Checks before every commit: `test -z "$(gofmt -l .)" && go vet ./... && go test ./...` (all three must be clean; the `test -z` makes a non-empty gofmt list fail)
- Integration tests (need a running daemon): `DOCKDU_INTEGRATION=1 go test ./...`
- Local release dry run: `goreleaser release --snapshot --clean`
- Demo GIF: `vhs demo.tape`

## Toolchain and dependencies

- Go 1.26. Modules: `github.com/moby/moby/client` v0.6.2 and `github.com/moby/moby/api` v1.56.1 only.
- IMPORTANT: never import `github.com/docker/docker`. It is deprecated and has open vulnerability advisories.
- Before using any client function, run `go doc github.com/moby/moby/client <Symbol>` and use the signature it prints. Do not rely on memory of older Docker SDKs.
- Use `client.New(client.FromEnv)`. `NewClientWithOpts` and `WithAPIVersionNegotiation` are deprecated.

## Code rules that differ from defaults

- Standard library first. A new third-party module needs a sentence in the commit message explaining why.
- Errors wrapped with `%w`; no panics outside `cmd/dockdu/main.go`.
- Every collector returns a partial result plus warnings instead of failing the whole report.
- dockdu never executes a state-changing command. Suggested commands are strings in `internal/report/suggest.go`.
- Never suggest `docker volume prune` or any `--volumes` flag.
- Output goes through `internal/report`; collectors do not print.

## Tests

- Unit tests use JSON fixtures under `internal/collect/testdata/`, recorded from a real daemon with the recording helper in `internal/engine`.
- Integration tests are gated by `DOCKDU_INTEGRATION=1` and `t.Skip` when the daemon is unreachable.
- Tests must not be weakened or deleted to make a build green. Fix the code.

## Git

- Conventional commits (`feat:`, `fix:`, `test:`, `docs:`, `chore:`), one logical change per commit.
- No `Co-Authored-By`, no "Generated with" or other AI attribution lines in commits, PRs or issues.
- Work per topic: issue, branch, PR, squash-merge. Pushing feature branches and merging PRs is authorised.
- Never tag, force-push, reset, delete branches or publish a release without asking.
- Commit identity: `Maxi <223365242+maximilianfeix@users.noreply.github.com>`.

## When compacting

Preserve: the list of modified files, the current phase from PROGRESS.md, open spike answers in docs/SPIKE.md, and the three check commands above.
