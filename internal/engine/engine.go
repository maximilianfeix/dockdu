// Package engine is the only place that talks to the Docker daemon. It
// produces a Snapshot, a plain, JSON-serialisable copy of everything dockdu
// needs, so that collectors and renderers never depend on the SDK and can
// be tested from recorded fixtures.
//
// The engine is strictly read-only: it calls Info, DiskUsage and
// ContainerInspect and nothing else.
package engine

import (
	"context"
	"errors"
	"time"
)

// ErrDaemonUnreachable wraps connection failures so main can map them to
// exit code 1.
var ErrDaemonUnreachable = errors.New("docker daemon unreachable")

// Engine returns a point-in-time Snapshot of the daemon's disk usage.
type Engine interface {
	Snapshot(ctx context.Context) (*Snapshot, error)
}

// Snapshot is everything dockdu reads from the daemon.
type Snapshot struct {
	TakenAt    time.Time     `json:"taken_at"`
	Daemon     Daemon        `json:"daemon"`
	Totals     Totals        `json:"totals"`
	Images     []Image       `json:"images"`
	Containers []Container   `json:"containers"`
	Volumes    []Volume      `json:"volumes"`
	BuildCache []CacheRecord `json:"build_cache"`
	// Warnings are non-fatal problems met while taking the snapshot, for
	// example a container that could not be inspected.
	Warnings []string `json:"warnings,omitempty"`
}

// Daemon describes the daemon the snapshot was taken from.
type Daemon struct {
	Host            string `json:"host"`
	ServerVersion   string `json:"server_version"`
	APIVersion      string `json:"api_version"`
	OperatingSystem string `json:"operating_system"`
	StorageDriver   string `json:"storage_driver"`
	LoggingDriver   string `json:"logging_driver"`
	DataRoot        string `json:"data_root"`
	// ContainerdSnapshotter is true when the daemon keeps images in the
	// containerd image store (DriverStatus driver-type
	// io.containerd.snapshotter.v1) instead of a graph driver.
	ContainerdSnapshotter bool `json:"containerd_snapshotter,omitempty"`
	// Rootless is true when Info.SecurityOptions contains name=rootless.
	Rootless bool `json:"rootless,omitempty"`
}

// Totals are the per-type aggregates the daemon computes for
// `docker system df`. They are reported, never recomputed.
type Totals struct {
	Images     Aggregate `json:"images"`
	Containers Aggregate `json:"containers"`
	Volumes    Aggregate `json:"volumes"`
	BuildCache Aggregate `json:"build_cache"`
}

// Aggregate mirrors the daemon's per-type disk usage summary.
type Aggregate struct {
	ActiveCount int64 `json:"active_count"`
	TotalCount  int64 `json:"total_count"`
	// Reclaimable is the daemon's own estimate of what a prune of this type
	// would free. For build cache it excludes shared records.
	Reclaimable int64 `json:"reclaimable"`
	TotalSize   int64 `json:"total_size"`
}

// Image is one entry of the image list. SharedSize and Containers are -1
// when the daemon did not compute them.
type Image struct {
	ID         string            `json:"id"`
	RepoTags   []string          `json:"repo_tags"`
	Created    int64             `json:"created"`
	Size       int64             `json:"size"`
	SharedSize int64             `json:"shared_size"`
	Containers int64             `json:"containers"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// Container is one entry of the container list plus the logging facts from
// inspect. LogPath is empty for drivers that do not report one (local).
type Container struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	ImageID    string            `json:"image_id"`
	Created    int64             `json:"created"`
	State      string            `json:"state"`
	Status     string            `json:"status"`
	SizeRw     int64             `json:"size_rw"`
	SizeRootFs int64             `json:"size_root_fs"`
	Labels     map[string]string `json:"labels,omitempty"`
	LogDriver  string            `json:"log_driver"`
	LogOpts    map[string]string `json:"log_opts,omitempty"`
	LogPath    string            `json:"log_path"`
	// InspectError is set when ContainerInspect failed; logging facts are
	// then unknown rather than empty.
	InspectError string `json:"inspect_error,omitempty"`
}

// Volume is one entry of the volume list. Size and RefCount are -1 when
// unknown; HasUsageData is false when the daemon sent no usage block at all.
type Volume struct {
	Name         string            `json:"name"`
	Driver       string            `json:"driver"`
	Mountpoint   string            `json:"mountpoint"`
	CreatedAt    string            `json:"created_at"`
	Labels       map[string]string `json:"labels,omitempty"`
	HasUsageData bool              `json:"has_usage_data"`
	Size         int64             `json:"size"`
	RefCount     int64             `json:"ref_count"`
}

// CacheRecord is one BuildKit cache record as returned by the daemon.
type CacheRecord struct {
	ID          string     `json:"id"`
	Parents     []string   `json:"parents,omitempty"`
	Type        string     `json:"type"`
	Description string     `json:"description"`
	InUse       bool       `json:"in_use"`
	Shared      bool       `json:"shared"`
	Size        int64      `json:"size"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	UsageCount  int        `json:"usage_count"`
}
