// Package collect turns an engine.Snapshot into report sections. Each
// collector returns the section plus warnings; none of them talks to the
// daemon or the filesystem directly except logs.go, which reads file sizes.
package collect

import (
	"github.com/maximilianfeix/dockdu/internal/engine"
	"github.com/maximilianfeix/dockdu/internal/report"
)

// Daemon copies the daemon facts into the report.
func Daemon(s *engine.Snapshot) report.Daemon {
	return report.Daemon{
		Host: s.Daemon.Host, ServerVersion: s.Daemon.ServerVersion, APIVersion: s.Daemon.APIVersion,
		OperatingSystem: s.Daemon.OperatingSystem, StorageDriver: s.Daemon.StorageDriver,
		LoggingDriver: s.Daemon.LoggingDriver, DataRoot: s.Daemon.DataRoot,
	}
}

// DfTotals fills the `docker system df` part of the summary from the
// daemon's own aggregates. Log bytes and the derived totals are added by
// Summarize once logs are measured.
func DfTotals(s *engine.Snapshot) report.Summary {
	sum := report.Summary{
		DfImages:     s.Totals.Images.TotalSize,
		DfContainers: s.Totals.Containers.TotalSize,
		DfVolumes:    s.Totals.Volumes.TotalSize,
		DfBuildCache: s.Totals.BuildCache.TotalSize,
	}
	sum.DfReportedBytes = sum.DfImages + sum.DfContainers + sum.DfVolumes + sum.DfBuildCache
	return sum
}

// Summarize combines df totals with measured log bytes. Shared build-cache
// bytes are already inside the image sizes, so they are subtracted once:
// total = df − shared build cache + logs. "Not shown by df" is the log bytes.
func Summarize(df report.Summary, buildCacheSharedBytes int64, logs report.Logs) report.Summary {
	df.LogBytes = logs.TotalBytes
	df.LogBytesComplete = logs.Complete
	df.BuildCacheSharedBytes = buildCacheSharedBytes
	df.DockerBytesTotal = df.DfReportedBytes - buildCacheSharedBytes + logs.TotalBytes
	df.NotShownByDf = logs.TotalBytes
	if df.Suggestions == nil {
		df.Suggestions = []string{}
	}
	return df
}
