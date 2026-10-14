# Security and read-only contract review

Date: 2026-10-14 · Branch `v0.2-hardening`

## 1. The contract

dockdu never changes Docker state. v0.1 kept this promise by convention: the engine package called only `Info`, `DiskUsage` and `ContainerInspect`. Nothing *enforced* it. A future change, a dependency update, or a contributor could have added a mutating call, and no test would have failed.

**Important:** mounting the socket as `-v /var/run/docker.sock:/var/run/docker.sock:ro` does **not** make the Docker API read-only. `:ro` only stops the container from replacing the socket file. Any process that can connect to the socket can send `POST`/`DELETE` requests and is effectively root on the host. The README and SECURITY.md now say so.

## 2. What enforces it now

| Layer | Mechanism | Test |
|---|---|---|
| Client request hook | `internal/engine/guard.go` `readOnlyGuard`, installed by `NewMoby` on every client. Refuses any method other than `GET`/`HEAD` and any path other than `/_ping`, `/v1.xx/info`, `/v1.xx/system/df`, `/v1.xx/containers/<12–64 hex>/json`. The hook runs inside moby/client's transport **before** the request is written to the socket. | `TestReadOnlyGuardTable` (13 cases, including `..` and `/logs`, `/archive`) |
| Regression test against a fake Engine | `fakeengine_test.go`: an `httptest` server records every request. (a) A full snapshot of 25 containers uses only allowlisted requests. (b) 13 client calls fail with `ErrNotReadOnly` and never reach the server: ContainerRemove, Stop, Kill, Create, ExecCreate, ImageRemove, VolumeRemove, Container/Image/Volume/BuildCache prune, ContainerLogs, CopyFromContainer. | `TestSnapshotUsesOnlyReadOnlyEndpoints`, `TestGuardRefusesMutatingCalls` |
| No command execution | dockdu runs no subprocess (`os/exec` is not imported anywhere; check with `grep -rn os/exec --include=*.go .`). Suggestions are strings. | code review |

Adding a mutating call now fails in two places: the guard refuses it at runtime, and the fake-engine tests fail in CI.

**Limits of the guard (by design):**
- It protects against dockdu's own code, not against a modified binary. A user who does not trust the binary should build from source or put a filtering proxy in front of the socket (§5).
- `GET /containers/<id>/json` returns the container's environment variables to dockdu's memory. dockdu copies only the log path and log config out of it and never prints or stores the rest.

## 3. Audit of every input

| Input | Trust | Handling | Status |
|---|---|---|---|
| Docker API responses (names, labels, tags, log options, IDs) | untrusted (remote or compromised daemon) | Text output passes every daemon string through `safe()`, which replaces control characters with `?`. Container IDs must match `^[0-9a-f]{12,64}$` before they are joined into a path. | v0.1 ✔ |
| Volume names inside suggested commands | untrusted | **Fixed:** shell-quoted, `--` before a leading dash, and no command at all when the name contains control characters. Every suggestion line also passes through `safe()`. Test: `TestSuggestVolumeRmQuotesDaemonNames`. | fixed |
| `LogPath` from inspect | untrusted | Must stay under the data root after `filepath.Clean` (`underRoot`). Only names `<base>`, `<base>.N` and `<base>.N.gz` are counted, via `lstat`, and only regular files. Contents are never opened. | v0.1 ✔, see S3 |
| Docker CLI config and context metadata (`~/.docker/config.json`, `contexts/meta/<sha>/meta.json`) | user-owned | Read with a 1 MiB limit (`readLimited`). Only `currentContext` and `Endpoints.docker.{Host,SkipTLSVerify}` are used. TLS files are passed by path to the client. `SkipTLSVerify` contexts are refused. `ssh://` is refused because it would need the docker CLI. | new |
| JSON size from the daemon | untrusted | `encoding/json` decodes into typed structs. A huge response costs memory proportional to its size; there is no explicit limit (S4). The 90 s timeout bounds slow responses. | open S4 |
| `--data-root` | user | Used as the root for log paths; nothing is written. | ✔ |
| Fixture recording (`DOCKDU_RECORD_FIXTURE`) | developer | **Fixed:** created with `O_EXCL`, so an existing file or symlink is never replaced. Sanitizes labels, registry hosts and build-step text. | fixed |

**Secrets:**
- dockdu prints no environment variables, credentials or file contents.
- `HostConfig.LogConfig.Config` (log options) is printed only as `max-size`/`max-file`.
- Fixtures drop non-compose labels.

**Network:** the only connection is the configured Docker endpoint. Nothing goes to the internet; there is no telemetry and no update check.

**Filesystem TOCTOU:**
- `ReadDir` then `Lstat` per file: a file swapped for a symlink between the two calls is detected by `Lstat` and skipped.
- A *directory* component replaced by a symlink between `underRoot` and `ReadDir` would make dockdu list a directory outside the root. Even then it only `lstat`s files named like the container's log file and reports their sizes, never their contents (S3).

## 4. Open security items

| ID | Sev | Finding | Proposed resolution | Verify |
|---|---|---|---|---|
| S1 | P2 | The container image runs as root (needed to read `root:root 0640` log files). | Keep root, but document `--read-only --cap-drop ALL --security-opt no-new-privileges` in the one-liner once tested. | Run the one-liner with those flags on Linux and Desktop. |
| S2 | P2 | Release archives are checksummed but not signed. There is no SBOM or provenance. | GoReleaser `signs` with cosign keyless and `sboms` with syft. Needs a release-workflow change (approval). | `cosign verify-blob` on a snapshot. |
| S3 | P3 | `ReadDir` follows symlinked directory components; a hostile `LogPath` under the root can name any file prefix in that directory. | Open the containers directory with `openat`-style walking (`os.Root`, Go ≥ 1.24) and refuse symlinked components. | Unit test with a symlinked `containers/<id>`. |
| S4 | P3 | No upper bound on daemon response size. | Wrap the transport body in a limit (for example 512 MiB) or accept as is: the user chose the daemon. | Fake engine sending a 1 GiB body. |
| S5 | P3 | `GET /containers/<id>/json` brings env vars into memory. | Accept: there is no narrower endpoint for the log config. | n/a |

## 5. Restricted socket proxy (optional hardening, not tested)

For people who will not hand the full socket to a third-party binary: [tecnativa/docker-socket-proxy](https://github.com/Tecnativa/docker-socket-proxy) filters the API by section and method. Per its README, `POST` is denied unless `POST=1`. dockdu needs the `INFO`, `SYSTEM` and `CONTAINERS` sections plus `/_ping`. An untested sketch:

```
docker network create dockdu-proxy
docker run -d --name dockdu-socket-proxy --network dockdu-proxy \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -e INFO=1 -e SYSTEM=1 -e CONTAINERS=1 -e POST=0 \
  tecnativa/docker-socket-proxy
docker run --rm --network dockdu-proxy -e DOCKER_HOST=tcp://dockdu-socket-proxy:2375 \
  -v /var/lib/docker:/var/lib/docker:ro ghcr.io/maximilianfeix/dockdu
```

Status: **NOT RUN**. Before documenting it in the README, verify that:
- `/system/df` is covered by `SYSTEM=1`;
- the proxy does not need `VERSION=1` for `/_ping`.

It adds a long-running container, so it stays an opt-in recipe rather than the default.

## 6. How to verify

```
go test ./internal/engine -run 'Guard|ReadOnly|Snapshot' -v
grep -rn '"os/exec"' --include='*.go' . ; echo "exit $? (1 = not imported)"
```
