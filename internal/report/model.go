// Package report holds the data model dockdu renders. Every byte value is an
// int64 in bytes; -1 means unknown. Each section carries its own
// suggestions (commands printed as text, never executed) and the warnings
// that explain which parts of it are partial.
package report

import "time"

// Report is the whole document `dockdu --json` serialises.
type Report struct {
	Version    string     `json:"version"`
	TakenAt    time.Time  `json:"taken_at"`
	Daemon     Daemon     `json:"daemon"`
	Summary    Summary    `json:"summary"`
	BuildCache BuildCache `json:"build_cache"`
	Logs       Logs       `json:"logs"`
	Projects   Projects   `json:"projects"`
	Images     Images     `json:"images"`
	Volumes    Volumes    `json:"volumes"`
	Desktop    *Desktop   `json:"desktop,omitempty"`
	Warnings   []string   `json:"warnings"`
}

// Daemon identifies where the numbers came from.
type Daemon struct {
	Host            string `json:"host"`
	ServerVersion   string `json:"server_version"`
	APIVersion      string `json:"api_version"`
	OperatingSystem string `json:"operating_system"`
	StorageDriver   string `json:"storage_driver"`
	LoggingDriver   string `json:"logging_driver"`
	DataRoot        string `json:"data_root"`
}

// Summary is the headline: what Docker reports, what dockdu measured on top,
// and the gap between the two.
type Summary struct {
	// DfReportedBytes is the sum of the four SIZE columns of `docker system df`.
	DfReportedBytes int64 `json:"df_reported_bytes"`
	DfImages        int64 `json:"df_images_bytes"`
	DfContainers    int64 `json:"df_containers_bytes"`
	DfVolumes       int64 `json:"df_volumes_bytes"`
	DfBuildCache    int64 `json:"df_build_cache_bytes"`
	// LogBytes is measured from files and is not part of df.
	LogBytes int64 `json:"log_bytes"`
	// LogBytesComplete is false when at least one container's logs could not
	// be measured; LogBytes is then a lower bound.
	LogBytesComplete bool `json:"log_bytes_complete"`
	// BuildCacheSharedBytes are build-cache bytes that also belong to an
	// image; df counts them in both rows, dockdu subtracts them once.
	BuildCacheSharedBytes int64 `json:"build_cache_shared_bytes"`
	// DockerBytesTotal = DfReportedBytes − BuildCacheSharedBytes + LogBytes.
	// Shared layers are counted once, as df does: a resource total, not a
	// disk-block count.
	DockerBytesTotal int64 `json:"docker_bytes_total"`
	// NotShownByDf is the log bytes: the only part df does not report.
	NotShownByDf int64    `json:"not_shown_by_df_bytes"`
	Suggestions  []string `json:"suggestions"`
}

// BuildCache follows SPEC §6.4.
type BuildCache struct {
	Records          int      `json:"records"`
	TotalBytes       int64    `json:"total_bytes"`
	SharedBytes      int64    `json:"shared_bytes"`
	PrivateBytes     int64    `json:"private_bytes"`
	ReclaimableBytes int64    `json:"reclaimable_bytes"`
	NotFreedByPrune  int64    `json:"not_freed_by_prune_bytes"`
	InUseBytes       int64    `json:"in_use_bytes"`
	Suggestions      []string `json:"suggestions"`
}

// Logs lists container log files on disk.
type Logs struct {
	TotalBytes  int64     `json:"total_bytes"`
	Complete    bool      `json:"complete"`
	Items       []LogItem `json:"items"`
	Suggestions []string  `json:"suggestions"`
}

// LogItem is one container's log footprint.
type LogItem struct {
	ContainerID string `json:"container_id"`
	Name        string `json:"name"`
	Project     string `json:"project,omitempty"`
	Driver      string `json:"driver"`
	// Rotation is "unbounded", "bounded" or "not on disk".
	Rotation string `json:"rotation"`
	MaxSize  string `json:"max_size,omitempty"`
	MaxFile  string `json:"max_file,omitempty"`
	Bytes    int64  `json:"bytes"`
	Files    int    `json:"files"`
	// Measured is false when the files could not be read; Bytes is then 0
	// and must not be summed.
	Measured bool   `json:"measured"`
	Note     string `json:"note,omitempty"`
}

// Projects groups resources by compose project. Images used by more than one
// project appear in each project's SharedImageBytes and are excluded from the
// project's TotalBytes; the section TotalBytes counts them once:
// TotalBytes == Σ Items.TotalBytes + SharedImageBytes.
type Projects struct {
	Items            []Project `json:"items"`
	TotalBytes       int64     `json:"total_bytes"`
	SharedImageBytes int64     `json:"shared_image_bytes"`
	Suggestions      []string  `json:"suggestions"`
}

// Project is one compose project (or the "(none)" bucket).
type Project struct {
	Name       string `json:"name"`
	Containers int    `json:"containers"`
	// ImageBytes counts images used only by this project.
	ImageBytes int64 `json:"image_bytes"`
	// SharedImageBytes counts images this project uses that other projects use
	// too. They are not part of TotalBytes, so project totals stay additive.
	SharedImageBytes int64 `json:"shared_image_bytes"`
	ImagesShared     int   `json:"images_shared_with_other_projects"`
	VolumeBytes      int64 `json:"volume_bytes"`
	VolumesUnknown   int   `json:"volumes_size_unknown"`
	LogBytes         int64 `json:"log_bytes"`
	TotalBytes       int64 `json:"total_bytes"`
	OrphanVolumes    int   `json:"orphan_volumes"`
	DanglingImages   int   `json:"dangling_images"`
	// StoppedOld counts non-running containers created more than 7 days ago
	// (the API has no stop time).
	StoppedOld int `json:"stopped_containers_created_over_7d"`
}

// Images follows SPEC §6.7.
type Images struct {
	Count         int   `json:"count"`
	TotalBytes    int64 `json:"total_bytes"`
	Dangling      int   `json:"dangling"`
	DanglingBytes int64 `json:"dangling_bytes"`
	// DanglingUnknown counts dangling images whose unique size is unknown; DanglingBytes is then a lower bound.
	DanglingUnknown  int         `json:"dangling_unknown"`
	ReclaimableBytes int64       `json:"reclaimable_bytes"`
	Items            []ImageItem `json:"items"`
	Suggestions      []string    `json:"suggestions"`
}

// ImageItem is one image with its unique-size rule applied.
type ImageItem struct {
	ID          string   `json:"id"`
	Tags        []string `json:"tags"`
	Bytes       int64    `json:"bytes"`
	SharedBytes int64    `json:"shared_bytes"`
	Containers  int64    `json:"containers"`
	Dangling    bool     `json:"dangling"`
	// UniqueBytes is Size-SharedSize when Containers==0 and both known; -1 otherwise.
	UniqueBytes int64    `json:"unique_bytes_if_removed"`
	Projects    []string `json:"projects,omitempty"`
}

// Volumes follows SPEC §6.8.
type Volumes struct {
	Count        int          `json:"count"`
	TotalBytes   int64        `json:"total_bytes"`
	UnknownSizes int          `json:"unknown_sizes"`
	Items        []VolumeItem `json:"items"`
	Suggestions  []string     `json:"suggestions"`
}

// VolumeItem is one volume.
type VolumeItem struct {
	Name     string `json:"name"`
	Driver   string `json:"driver"`
	Project  string `json:"project,omitempty"`
	Bytes    int64  `json:"bytes"`
	RefCount int64  `json:"ref_count"`
	// Reference is "referenced", "unreferenced" or "unknown".
	Reference string `json:"reference"`
	Orphan    bool   `json:"orphan"`
}

// Desktop is the macOS Docker Desktop / OrbStack section.
type Desktop struct {
	Product        string `json:"product"`
	DiskFile       string `json:"disk_file"`
	ApparentBytes  int64  `json:"apparent_bytes"`
	AllocatedBytes int64  `json:"allocated_bytes"`
	// DockerReportedBytes is Summary.DockerBytesTotal, repeated here next to
	// the allocated size. The two are different measures and may overlap.
	DockerReportedBytes int64  `json:"docker_reported_bytes"`
	Explanation         string `json:"explanation"`
}
