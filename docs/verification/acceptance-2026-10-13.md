# Acceptance run, 2026-10-13

SPEC.md §9, run on commit 05edab3 (branch phase-7-release) on macOS arm64, Docker Desktop (engine 29.2.0, API 1.53), Go 1.27.2, GoReleaser 2.18.3. Outputs are trimmed only where marked.

| # | Criterion | Result |
|---|---|---|
| 1 | `go build ./... && go vet ./... && go test ./...` exit 0 | PASS |
| 2 | `gofmt -l .` prints nothing | PASS |
| 3 | Unit tests cover build-cache arithmetic, log paths, compose grouping incl. shared images, image rules, dangling, renderers, from fixtures | PASS (49 tests, see inventory) |
| 4 | `DOCKDU_INTEGRATION=1` test creates a 1k×3 container and asserts the rotated footprint | PASS inside `golang:1.27` with the socket and `/var/lib/docker:ro` mounted (`make integration`), locally and in the CI `integration` job on ubuntu-latest; SKIP on the macOS host because the data root is inside the VM |
| 5 | `./dockdu --json \| jq -e '.summary.docker_bytes_total > 0'` | PASS |
| 6 | No daemon: exit 1, one line naming `DOCKER_HOST` and the socket | PASS |
| 7 | Container image prints the same report as the local binary on the same host | PASS with one documented difference: on Docker Desktop only the container run can measure logs (host: `logs_complete=false`, 7 warnings; container: `logs_complete=true`, 8,026,007 log bytes). df, build cache, image and volume numbers agree; df differs by the running container's own 70 KB rw layer |
| 8 | README has `demo.gif` from `vhs demo.tape`, install commands, "what it shows", "never deletes anything" | PASS |
| 9 | `goreleaser release --snapshot --clean` succeeds | PASS (4 archives; with the docker stage on the project builder both images built, 3.4 MB amd64 / 10.6 MB arm64) |

Not run (need a published release): Homebrew install from the tap, pulling the image from ghcr.io.

## 1–2

```
$ go build ./... && go vet ./... && go test ./...
ok  	github.com/maximilianfeix/dockdu/cmd/dockdu
ok  	github.com/maximilianfeix/dockdu/internal/collect
ok  	github.com/maximilianfeix/dockdu/internal/engine
ok  	github.com/maximilianfeix/dockdu/internal/report
exit=0
$ gofmt -l .
(empty)
```

## 3 test inventory

TestAttachImageProjects, TestBuildCacheDuplicateRecordsCountedOnce, TestBuildCacheEmpty, TestBuildCacheInvariantViolationIsWarned, TestBuildCacheNegativePrivateIsUnknown, TestBuildCacheSuggestions, TestBuildCacheTotalMismatchIsWarned, TestBuildCacheUsesEngineReclaimableNotBuildxRule, TestBytes, TestDfTotalsFromRecordedFixture, TestDfTotalsSumsTheFourDfRows, TestImagesDanglingAndUniqueRules, TestImagesDuplicateIDsAndSortTieBreak, TestImagesEmptyAndSorted, TestImagesSuggestionsFollowDangling, TestIntegrationLogsRotation, TestLogsDataRootOverrideRebasesReportedPath, TestLogsEmpty, TestLogsInspectErrorIsUnmeasured, TestLogsJSONFileSumsRotatedSiblingsAndFlagsRotation, TestLogsJSONFileUnboundedDerivesPathAndSuggests, TestLogsLocalDriverUsesLocalLogsDir, TestLogsMissingDirIsUnmeasuredNotZero, TestLogsOtherDriverIsNotOnDisk, TestLogsRejectsInvalidIDsAndPathsOutsideRoot, TestLogsSortedBySizeDescThenName, TestLogsSymlinksAreNotFollowed, TestProjectsEmpty, TestProjectsEmptyImageIDFallsBackToTag, TestProjectsGroupAndCountSharedImagesOnce, TestRedactHelpers, TestRenderExitZeroWhenEverythingIsMeasured, TestRenderFromFixtureExitCodes, TestRunHelp, TestRunRejectsPositionalArguments, TestRunUnknownFlag, TestRunUnreachableDaemonExits1, TestRunVersion, TestSanitizeDropsPrivateData, TestStaticReturnsCopy, TestSummarizeAddsLogsAndFlagsIncomplete, TestVolumesDuplicateNamesCountedOnce, TestVolumesReferenceStatesAndOrphans, TestWriteJSONListsAreNeverNull, TestWriteLoadRoundTripAndNoOverwrite, TestWriteTextGolden, TestWriteTextHasNoTrailingSpaces, TestWriteTextNeutralisesControlCharacters, TestWriteTextNeverSuggestsBulkVolumeRemoval, TestWriteTextTopLimitsRowsAndSaysSo

## 4

```
$ make integration
=== RUN   TestIntegrationLogsRotation
    logs_integration_test.go:98: measured 3116 bytes in 3 files for caf970d0443a
--- PASS: TestIntegrationLogsRotation (4.23s)
ok  	github.com/maximilianfeix/dockdu/internal/collect	4.249s
```

The test created one container labelled `dockdu.test=1` through the SDK and removed it by ID in cleanup. The same target passed in GitHub Actions (job `integration`, run 38442592663).

## 5–6

```
$ ./dockdu --json | jq -e '.summary.docker_bytes_total > 0'
true
$ DOCKER_HOST=unix:///nonexistent/dockdu.sock ./dockdu
dockdu: cannot reach the Docker daemon at unix:///nonexistent/dockdu.sock
  set DOCKER_HOST or start Docker; detail: docker daemon unreachable: info: failed to connect to the docker API at unix:///nonexistent/dockdu.sock; … (trimmed)
exit=1
```

## 7

```
host binary:     {"df":7247352014,"bc":1546143746,"images":15,"volumes":3,"logs_complete":false,"warnings":7}
container image: {"df":7247421646,"bc":1546143746,"images":15,"volumes":3,"logs_complete":true,"warnings":0,"log_bytes":8026007}
```

## 8

`docs/demo.gif` 355,093 bytes, 1400×1040, recorded with `vhs demo.tape` against the labelled demo compose project `dockdu-demo` and the orphan volume `dockdu-orphan-demo`. README contains the install commands, the "What it shows" section and the sentence "dockdu itself never removes, stops or prunes anything".

## 9

```
$ goreleaser check
  • 1 configuration file(s) validated
$ goreleaser release --snapshot --clean
  • release succeeded after 9s
$ docker images ghcr.io/maximilianfeix/dockdu
ghcr.io/maximilianfeix/dockdu:v0.0.0-SNAPSHOT-bd1e9df-amd64 3.4MB
ghcr.io/maximilianfeix/dockdu:v0.0.0-SNAPSHOT-bd1e9df-arm64 10.6MB
```

The generated cask (`dist/homebrew/Casks/dockdu.rb`) contains the `postflight` block with `system_command "/usr/bin/xattr"` and per-platform URLs and checksums.
