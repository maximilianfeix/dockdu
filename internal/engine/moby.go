package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/moby/moby/client"
)

// Moby is the Engine backed by the official Docker client.
type Moby struct {
	cli *client.Client
}

// inspectWorkers bounds concurrent ContainerInspect calls. Each inspect is
// a cheap metadata read; a small pool hides socket round trips on hosts with
// many containers without flooding the daemon. A variable only so the
// benchmark can compare against sequential inspection.
var inspectWorkers = 4

// NewMoby connects using the standard DOCKER_* environment variables; opts
// (for example from ResolveTarget) are applied after them. API version
// negotiation is the client's default. Every request passes readOnlyGuard,
// which refuses anything but GET/HEAD on the four endpoints dockdu reads.
// Connecting is lazy; the first call reports ErrDaemonUnreachable when the
// socket is not there.
func NewMoby(opts ...client.Opt) (*Moby, error) {
	all := append([]client.Opt{client.FromEnv}, opts...)
	all = append(all, client.WithHTTPRequestHook(readOnlyGuard))
	cli, err := client.New(all...)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	return &Moby{cli: cli}, nil
}

// Host is the daemon address the client will use.
func (m *Moby) Host() string { return m.cli.DaemonHost() }

// Close releases the underlying HTTP client.
func (m *Moby) Close() error { return m.cli.Close() }

// Snapshot performs Info, DiskUsage and one ContainerInspect per container.
// A failed inspect does not fail the snapshot; it is recorded on the
// container and in Warnings.
func (m *Moby) Snapshot(ctx context.Context) (*Snapshot, error) {
	infoRes, err := m.cli.Info(ctx, client.InfoOptions{})
	if err != nil {
		return nil, wrapConn("info", err)
	}
	du, err := m.cli.DiskUsage(ctx, client.DiskUsageOptions{
		Containers: true, Images: true, Volumes: true, BuildCache: true, Verbose: true,
	})
	if err != nil {
		return nil, wrapConn("disk usage", err)
	}

	s := &Snapshot{
		TakenAt:    time.Now().UTC(),
		Images:     []Image{},
		Containers: []Container{},
		Volumes:    []Volume{},
		BuildCache: []CacheRecord{},
		Daemon: Daemon{
			Host:                  m.cli.DaemonHost(),
			ServerVersion:         infoRes.Info.ServerVersion,
			APIVersion:            m.cli.ClientVersion(),
			OperatingSystem:       infoRes.Info.OperatingSystem,
			StorageDriver:         infoRes.Info.Driver,
			LoggingDriver:         infoRes.Info.LoggingDriver,
			DataRoot:              infoRes.Info.DockerRootDir,
			ContainerdSnapshotter: containerdSnapshotter(infoRes.Info.DriverStatus),
			Rootless:              rootless(infoRes.Info.SecurityOptions),
		},
		Totals: Totals{
			Images:     Aggregate{du.Images.ActiveCount, du.Images.TotalCount, du.Images.Reclaimable, du.Images.TotalSize},
			Containers: Aggregate{du.Containers.ActiveCount, du.Containers.TotalCount, du.Containers.Reclaimable, du.Containers.TotalSize},
			Volumes:    Aggregate{du.Volumes.ActiveCount, du.Volumes.TotalCount, du.Volumes.Reclaimable, du.Volumes.TotalSize},
			BuildCache: Aggregate{du.BuildCache.ActiveCount, du.BuildCache.TotalCount, du.BuildCache.Reclaimable, du.BuildCache.TotalSize},
		},
	}

	for _, im := range du.Images.Items {
		s.Images = append(s.Images, Image{
			ID: im.ID, RepoTags: append([]string(nil), im.RepoTags...), Created: im.Created,
			Size: im.Size, SharedSize: im.SharedSize, Containers: im.Containers, Labels: copyMap(im.Labels),
		})
	}
	for _, v := range du.Volumes.Items {
		vol := Volume{Name: v.Name, Driver: v.Driver, Mountpoint: v.Mountpoint, CreatedAt: v.CreatedAt,
			Labels: copyMap(v.Labels), Size: -1, RefCount: -1}
		if v.UsageData != nil {
			vol.HasUsageData = true
			vol.Size = v.UsageData.Size
			vol.RefCount = v.UsageData.RefCount
		}
		s.Volumes = append(s.Volumes, vol)
	}
	for _, r := range du.BuildCache.Items {
		s.BuildCache = append(s.BuildCache, CacheRecord{
			ID: r.ID, Parents: append([]string(nil), r.Parents...), Type: r.Type, Description: r.Description,
			InUse: r.InUse, Shared: r.Shared, Size: r.Size, CreatedAt: r.CreatedAt, LastUsedAt: r.LastUsedAt,
			UsageCount: r.UsageCount,
		})
	}
	s.Containers = make([]Container, len(du.Containers.Items))
	for i, c := range du.Containers.Items {
		s.Containers[i] = Container{
			ID: c.ID, Name: primaryName(c.Names), Image: c.Image, ImageID: c.ImageID, Created: c.Created,
			State: string(c.State), Status: c.Status, SizeRw: c.SizeRw, SizeRootFs: c.SizeRootFs, Labels: copyMap(c.Labels),
		}
	}
	if err := m.inspectAll(ctx, s.Containers); err != nil {
		return nil, err
	}
	// Warnings in container order, independent of which worker finished first.
	for _, ct := range s.Containers {
		if ct.InspectError != "" {
			s.Warnings = append(s.Warnings, fmt.Sprintf("inspect %s (%s): %s", ct.Name, short(ct.ID), ct.InspectError))
		}
	}
	return s, nil
}

// inspectAll fills the logging facts of every container with at most
// inspectWorkers concurrent ContainerInspect calls. A failed inspect is
// recorded on the container; a lost connection or a cancelled context
// fails the whole snapshot.
func (m *Moby) inspectAll(ctx context.Context, cs []Container) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan int)
	var (
		wg       sync.WaitGroup
		once     sync.Once
		fatalErr error
	)
	fail := func(err error) {
		once.Do(func() { fatalErr = err; cancel() })
	}
	for range min(inspectWorkers, len(cs)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				ct := &cs[i]
				ins, err := m.cli.ContainerInspect(ctx, ct.ID, client.ContainerInspectOptions{})
				switch {
				case err == nil:
					ct.LogPath = ins.Container.LogPath
					if hc := ins.Container.HostConfig; hc != nil {
						ct.LogDriver = hc.LogConfig.Type
						ct.LogOpts = copyMap(hc.LogConfig.Config)
					}
				case client.IsErrConnectionFailed(err):
					fail(wrapConn("inspect", err))
				case ctx.Err() != nil:
					fail(ctx.Err())
				default:
					ct.InspectError = err.Error()
				}
			}
		}()
	}
feed:
	for i := range cs {
		select {
		case jobs <- i:
		case <-ctx.Done():
			break feed
		}
	}
	close(jobs)
	wg.Wait()
	if fatalErr != nil {
		return fatalErr
	}
	return ctx.Err()
}

// containerdSnapshotter reports whether DriverStatus names the containerd
// snapshotter, which is how the daemon says it uses the containerd image
// store.
func containerdSnapshotter(status [][2]string) bool {
	for _, kv := range status {
		if kv[0] == "driver-type" && kv[1] == "io.containerd.snapshotter.v1" {
			return true
		}
	}
	return false
}

// rootless reports whether the daemon runs in rootless mode, which it
// announces as the security option "name=rootless".
func rootless(securityOptions []string) bool {
	for _, o := range securityOptions {
		for _, kv := range strings.Split(o, ",") {
			if kv == "name=rootless" {
				return true
			}
		}
	}
	return false
}

func wrapConn(op string, err error) error {
	if client.IsErrConnectionFailed(err) {
		return fmt.Errorf("%w: %s: %v", ErrDaemonUnreachable, op, err)
	}
	return fmt.Errorf("%s: %w", op, err)
}

func primaryName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return strings.TrimPrefix(names[0], "/")
}

func short(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func copyMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
