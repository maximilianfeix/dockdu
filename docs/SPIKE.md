# Phase 0 spike: feasibility and contracts

Date: 2026-10-13. Every item below was run on this machine unless marked NOT RUN. "Observed" means runtime behaviour seen here; "source" means read from upstream code or docs without executing it. Container IDs are real but belong to throwaway test containers created for this spike (label `dockdu.test=1`). No credentials appear in this file.

## Environment

| Item | Value | Command |
|---|---|---|
| Host | macOS, Apple Silicon (arm64) | `uname -m` |
| Go | go1.27.2 darwin/arm64 (Homebrew bottle) | `go version` |
| GoReleaser | 2.18.3 | `goreleaser --version` |
| Docker client / server | 29.2.0 / 29.2.0 | `docker version` |
| Engine API | server 1.53, min 1.44, client 1.53 | `docker version --format '{{.Server.APIVersion}} {{.Server.MinAPIVersion}}'` |
| Context / endpoint | `desktop-linux`, `unix:///Users/<user>/.docker/run/docker.sock` | `docker context inspect --format '{{.Endpoints.docker.Host}}'` |
| Daemon | Docker Desktop, aarch64, storage driver `overlayfs`, containerd snapshotter enabled, logging driver `json-file`, data root `/var/lib/docker` | `docker info` |
| daemon.json | only `builder.gc` (`defaultKeepStorage: 20GB`) and `experimental: false`; no `log-driver`, no `log-opts` | `cat ~/.docker/daemon.json` |
| Desktop settings | `UseContainerdSnapshotter: true`; `DiskSizeMiB` and Enhanced Container Isolation keys absent (defaults) | `settings-store.json` (read only) |
| vhs / ttyd / ffmpeg / jq / gh | present, gh logged in as the repo owner | `which`, `gh auth status` |

Go 1.27.2 satisfies the spec's minimum of 1.26; `go.mod` will declare `go 1.26`.

## Q1. Does a read-only bind mount of /var/lib/docker inside a container on Docker Desktop (macOS) expose container log files and the buildkit directory?

Observed: yes.

Setup (test containers):

```
docker run -d --name dockdu-test-logs --label dockdu.test=1 --log-opt max-size=1k --log-opt max-file=3 alpine:3.22 sh -c '…400 lines…; sleep 3600'
docker run -d --name dockdu-test-default --label dockdu.test=1 alpine:3.22 sh -c 'echo dockdu default; sleep 3600'
```

Command:

```
docker run --rm -v /var/lib/docker:/var/lib/docker:ro alpine:3.22 sh -c 'ls -ld /var/lib/docker /var/lib/docker/containers /var/lib/docker/buildkit; for d in /var/lib/docker/containers/*/; do ls -l "$d" | grep -E "json.log|local-logs"; done'
```

Relevant output:

```
drwx--x---   13 root  root  4096 Oct 13 16:55 /var/lib/docker
drwx--x--x    5 root  root  4096 Oct  5 19:54 /var/lib/docker/buildkit
drwx--x---    5 root  root  4096 Oct 13 16:57 /var/lib/docker/containers
-rw-r-----  1047  1f1412f2177f…-json.log
-rw-r-----  1048  1f1412f2177f…-json.log.1
-rw-r-----  1042  1f1412f2177f…-json.log.2
-rw-r-----    85  871547ac7ec5…-json.log
```

Conclusion: the mount resolves inside the Desktop VM. The current log file and its rotated siblings `.1`, `.2` are listable and their sizes readable as root inside the container. Rotation with `max-size=1k max-file=3` produced exactly three files. The buildkit directory is present.

Second observation, same mount, on this machine: `/var/lib/docker/overlay2` and `/var/lib/docker/image` do not exist (`du: No such file or directory`). `du -sx /var/lib/docker` is 1.5 MiB and `/var/lib/containerd` is 4 KiB while `docker system df` reports 3.5 GB of images. Where the image bytes live was not determined (`-x` does not cross mount points, so a sub-mount under the data root would be invisible to this check). `df -h /var/lib/docker` inside the VM shows `/dev/vda1 223.6G total, 4.5G used`.

Remaining uncertainty: tested on one Docker Desktop version (29.2.0 engine, containerd snapshotter on). With Enhanced Container Isolation enabled, Docker documents that host-namespace sharing and socket mounts are restricted; not tested here. The ghcr image runs as root inside the container, which is why the `root:root 0640` log files are readable; a non-root image would not read them.

## Q2. Are daemon-level log-opts reflected in each container's HostConfig.LogConfig, or left empty?

Observed on this machine (no daemon log-opts configured):

```
docker inspect dockdu-test-logs dockdu-test-default --format '{{.Name}} driver={{.HostConfig.LogConfig.Type}} cfg={{json .HostConfig.LogConfig.Config}} logpath={{.LogPath}}'
/dockdu-test-logs     driver=json-file cfg={"max-file":"3","max-size":"1k"} logpath=/var/lib/docker/containers/1f14…/1f14…-json.log
/dockdu-test-default  driver=json-file cfg={}                               logpath=/var/lib/docker/containers/8715…/8715…-json.log
```

So per-container `--log-opt` values appear in `HostConfig.LogConfig.Config`, and a container created without options has an empty map when the daemon has no defaults.

Source (moby master, `daemon/logs.go`, `mergeAndVerifyLogConfig`): when the container's log type is empty it is set to the daemon default; when the container's type equals the daemon default type, every daemon `log-opts` key that the container did not set is copied into `cfg.Config`. This runs at container create time.

Conclusion: `HostConfig.LogConfig.Config` as returned by inspect contains the effective options, including inherited daemon defaults, for containers whose driver matches the daemon default. A container created with an explicit different driver does not inherit the daemon's opts.

NOT RUN: the runtime confirmation of inheritance, because this daemon has no `log-opts` and changing the user's daemon.json was not authorised. The conclusion rests on the source reading plus the observed empty map.

Implication for dockdu: "rotation bounded" = `Config["max-size"]` present, for `json-file`. For `local` the driver rotates by default (20m × 5) regardless of the map.

## Q2b. LogPath for the `local` driver

Observed:

```
docker run -d --name dockdu-test-local --label dockdu.test=1 --log-driver local alpine:3.22 sh -c '…50 lines…; sleep 3600'
docker inspect dockdu-test-local --format 'driver={{.HostConfig.LogConfig.Type}} cfg={{json .HostConfig.LogConfig.Config}} logpath=[{{.LogPath}}]'
driver=local cfg={} logpath=[]
```

Inside the VM (read-only mount): `/var/lib/docker/containers/<id>/local-logs/container.log` (2740 bytes), directory mode `0700`.

Conclusion: observed at runtime: `LogPath` is empty for `local` and the file is `<data-root>/containers/<id>/local-logs/container.log`; dockdu sums `container.log*` in that directory. Not observed (docs only, docs.docker.com/engine/logging/drivers/local/): the rotated naming `container.log.1.gz` and the defaults 20m × 5 files, compressed; the test file stayed at 2740 bytes without rotating.

## Q3. Concrete Go type of DiskUsageResult.BuildCache.Items

Command (scratch module with `go get github.com/moby/moby/client@v0.6.2 github.com/moby/moby/api@v1.56.1`):

```
go doc github.com/moby/moby/client BuildCacheDiskUsage
```

Output: `Items []build.CacheRecord`, plus `ActiveCount int64`, `TotalCount int64`, `Reclaimable int64`, `TotalSize int64`.

`build.CacheRecord` fields: `ID string`, `Parents []string`, `Type string`, `Description string`, `InUse bool`, `Shared bool`, `Size int64`, `CreatedAt time.Time`, `LastUsedAt *time.Time`, `UsageCount int`. No `Parent` field.

Conclusion: matches the spec. Value slice, not pointers.

## Q4. Signatures of Client.DiskUsage and Client.ContainerInspect

```
func (cli *Client) DiskUsage(ctx context.Context, options DiskUsageOptions) (DiskUsageResult, error)
type DiskUsageOptions struct { Containers, Images, BuildCache, Volumes, Verbose bool }
type DiskUsageResult struct { Containers ContainersDiskUsage; Images ImagesDiskUsage; BuildCache BuildCacheDiskUsage; Volumes VolumesDiskUsage }

func (cli *Client) ContainerInspect(ctx context.Context, containerID string, options ContainerInspectOptions) (ContainerInspectResult, error)
type ContainerInspectOptions struct { Size bool }   // Size is costly, leave false
type ContainerInspectResult struct { Container container.InspectResponse; Raw json.RawMessage }

func (cli *Client) Info(ctx context.Context, options InfoOptions) (SystemInfoResult, error)
type SystemInfoResult struct { Info system.Info }   // Info.DockerRootDir, .LoggingDriver, .ServerVersion, .OperatingSystem, .Driver

func New(ops ...Opt) (*Client, error)
func FromEnv(c *clientConfig) error
func WithHost(host string) Opt
func IsErrConnectionFailed(err error) bool
func (cli *Client) DaemonHost() string
```

Type fields confirmed with `go doc`: `image.Summary{ID, RepoTags, Size, SharedSize, Containers, Labels}` (no VirtualSize); `container.Summary{ID, Names, Image, ImageID, Created int64, SizeRw, SizeRootFs, Labels, State ContainerState, Status}`; `volume.Volume{Name, Driver, Labels, Mountpoint, CreatedAt, UsageData *UsageData{RefCount, Size}}`; `container.LogConfig{Type string; Config map[string]string}`; `container.InspectResponse{LogPath, Name, HostConfig *HostConfig, …}`.

Resolved dependency tree (`go.mod` of the scratch module): the two direct modules pull in 18 indirect modules: `github.com/docker/go-connections`, `github.com/docker/go-units`, `github.com/distribution/reference`, `github.com/containerd/errdefs`, `github.com/containerd/errdefs/pkg`, `github.com/opencontainers/{go-digest,image-spec}`, `github.com/moby/docker-image-spec`, `github.com/Microsoft/go-winio`, `golang.org/x/sys`, and the OpenTelemetry modules `go.opentelemetry.io/otel{,/metric,/trace}`, `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`, `go.opentelemetry.io/auto/sdk`, `github.com/go-logr/{logr,stdr}`, `github.com/felixge/httpsnoop`. These are transitive requirements of the SDK and are accepted as such.

Conclusion: the spec's API section is correct. `IsErrConnectionFailed` is the right check for exit code 1.

## Q4b. Engine `/system/df` behaviour on this daemon

```
docker system df
TYPE            TOTAL  ACTIVE  SIZE      RECLAIMABLE
Images          8      1       3.547GB   1.105GB (31%)
Containers      2      2       8.192kB   0B (0%)
Local Volumes   0      0       0B        0B
Build Cache     26     0       1.546GB   1.119GB

docker buildx du | tail -4
Shared:      427.3MB
Private:     1.119GB
Reclaimable: 1.546GB
Total:       1.546GB

curl --unix-socket … 'http://localhost/v1.52/system/df?type=build-cache&verbose=true'
{"BuildCacheUsage":{"TotalSize":1546143746,"Reclaimable":1118813063,"Items":[26 records with ID, Type, InUse, Shared, Size, CreatedAt, LastUsedAt, UsageCount]}}
```

Observed difference: the engine's `Reclaimable` (1.119 GB, what `docker system df` prints) equals `Total − Shared` here, while `docker buildx du` reports `Reclaimable` = 1.546 GB (every record is `!InUse`). Source confirms the rule beyond this one data point: moby `daemon/internal/builder-next/builder.go`, `Builder.DiskUsage`, adds a record to `Reclaimable` only `if !r.InUse && !r.Shared`, and every record to `TotalSize`. Invariant for tests: `Reclaimable + Σ Size(Shared || InUse) == TotalSize`. Note for implementers: `build.CacheRecord.Parents` carries the json tag `" Parents"` (leading space) in api v1.56.1 and never decodes; dockdu does not use it. The spec's 6.4 formula followed buildx. dockdu must show the engine number as "what `docker system df` shows" and use `Shared || InUse` for "prune would NOT free". Spec amendment proposed in the PR.

The API with `type=build-cache` returns the v1.52+ shape `BuildCacheUsage` (no legacy top-level arrays), consistent with the client's `DiskUsageResult`.

## Q5. GoReleaser syntax for homebrew_casks hooks in the installed version

Observed with GoReleaser 2.18.3 (`goreleaser check` on two scratch configs):

- `hooks.post.install_steps` → `yaml: unmarshal errors: line 11: field install_steps not found in type config.HomebrewCaskHook`.
- `hooks.post.install: |` with a Ruby string → parses; the only remaining error was "no remote configured" because the scratch repo has no origin.
- `goreleaser jsonschema` lists `HomebrewCaskHook` with properties `install` and `uninstall`, and `HomebrewCaskHooks` with `post` and `pre`. `dockers_v2` parsed without error.

Docs (goreleaser.com/customization/homebrew_casks/, read 2026-10-13): the `_steps` form is labelled "Since v2.19 (unreleased)"; `install` and `uninstall` are marked deprecated but remain supported and are still rendered when `_steps` is absent.

Conclusion: use the `install:` string form with the Homebrew cask DSL:

```yaml
hooks:
  post:
    install: |
      system_command "/usr/bin/xattr", args: ["-dr", "com.apple.quarantine", "#{staged_path}/dockdu"]
```

Spec section 10 amended accordingly. Full validation of the config against a real remote happens in Phase 7 with `goreleaser check`.

Remaining uncertainty: the generated cask was not installed from a tap here (publishing is out of scope for this phase), so the quarantine hook is validated syntactically, not by a Homebrew install.

## Docker Desktop VM disk file

```
F=~/Library/Containers/com.docker.docker/Data/vms/0/data/Docker.raw
stat -f 'apparent=%z allocated_blocks=%b' "$F"   → apparent=245106737152 allocated_blocks=9326448
du -h "$F"                                        → 4,4G
```

Apparent size 228.3 GiB (the default maximum, `DiskSizeMiB` unset), allocated 9326448 × 512 B = 4.45 GiB, which matches `df` inside the VM (4.5 G used). `syscall.Stat_t.Blocks` is in 512-byte units on macOS regardless of the 4096 block size reported by `%k`.

Conclusion: `os.Stat().Size()` vs `Stat_t.Blocks*512` is the right pair. Because image bytes are not under `/var/lib/docker` with the containerd snapshotter, dockdu will not try to `du` inside the VM (that would require running a container, which is a state change). The Desktop section compares the allocated file size with the sum of API-reported resource sizes and labels them as different measures. Spec 6.10 amended.

## Spec amendments proposed in this PR

1. 6.4: define `reclaimable_engine = Reclaimable from the API` (matches `docker system df`) and `not_freed_by_prune = Σ Size where Shared || InUse`; drop the buildx-style `!InUse` definition.
2. 6.10: replace "bytes used inside the VM (sum of sections above)" with "sum of Docker-reported resource sizes, shown next to the allocated file size with a note that the two are different measures and may overlap".
3. 10: `homebrew_casks.hooks.post.install` string form; note GoReleaser 2.18.3.
4. 6.1: note Go 1.27.2 installed, `go 1.26` in go.mod; list of transitive modules accepted.
5. 6.5: no change needed; both drivers confirmed at runtime.

## Test resources created in this phase

| Resource | ID (short) | Purpose |
|---|---|---|
| container `dockdu-test-logs` | 1f1412f2177f | json-file with 1k × 3 rotation |
| container `dockdu-test-default` | 871547ac7ec5 | json-file without options |
| container `dockdu-test-local` | 34b1d5c11f04 | local driver |
| image `alpine:3.22` | pulled | base for the above and for the integration test |

All three containers carry the label `dockdu.test=1` and are removed by ID at the end of the project, never by name prefix. The image stays for the integration test.
