package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/moby/moby/client"
)

// Moby is the Engine backed by the official Docker client.
type Moby struct {
	cli *client.Client
}

// NewMoby connects using the standard DOCKER_* environment variables.
// API version negotiation is the client's default. Connecting is lazy; the
// first call reports ErrDaemonUnreachable when the socket is not there.
func NewMoby(opts ...client.Opt) (*Moby, error) {
	all := append([]client.Opt{client.FromEnv}, opts...)
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
			Host:            m.cli.DaemonHost(),
			ServerVersion:   infoRes.Info.ServerVersion,
			APIVersion:      m.cli.ClientVersion(),
			OperatingSystem: infoRes.Info.OperatingSystem,
			StorageDriver:   infoRes.Info.Driver,
			LoggingDriver:   infoRes.Info.LoggingDriver,
			DataRoot:        infoRes.Info.DockerRootDir,
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
	for _, c := range du.Containers.Items {
		ct := Container{
			ID: c.ID, Name: primaryName(c.Names), Image: c.Image, ImageID: c.ImageID, Created: c.Created,
			State: string(c.State), Status: c.Status, SizeRw: c.SizeRw, SizeRootFs: c.SizeRootFs, Labels: copyMap(c.Labels),
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		ins, err := m.cli.ContainerInspect(ctx, c.ID, client.ContainerInspectOptions{})
		if err != nil {
			if client.IsErrConnectionFailed(err) {
				return nil, wrapConn("inspect", err)
			}
			ct.InspectError = err.Error()
			s.Warnings = append(s.Warnings, fmt.Sprintf("inspect %s (%s): %v", ct.Name, short(c.ID), err))
		} else {
			ct.LogPath = ins.Container.LogPath
			if hc := ins.Container.HostConfig; hc != nil {
				ct.LogDriver = hc.LogConfig.Type
				ct.LogOpts = copyMap(hc.LogConfig.Config)
			}
		}
		s.Containers = append(s.Containers, ct)
	}
	return s, nil
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
