package collect

import (
	"time"

	"github.com/maximilianfeix/dockdu/internal/engine"
	"github.com/maximilianfeix/dockdu/internal/report"
)

// Build runs every collector over a snapshot and assembles the report.
// dataRoot overrides the daemon's data root for log measurement ("" keeps
// the daemon value). Warnings from all collectors are gathered in order.
func Build(s *engine.Snapshot, dataRoot string, now time.Time, version string) *report.Report {
	r := &report.Report{Version: version, TakenAt: s.TakenAt, Daemon: Daemon(s), Warnings: []string{}}
	r.Warnings = append(r.Warnings, s.Warnings...)
	if dataRoot != "" && dataRoot != r.Daemon.DataRoot {
		r.Daemon.ReportedDataRoot = r.Daemon.DataRoot
		r.Daemon.DataRoot = dataRoot
	}

	var w []string
	r.BuildCache, w = BuildCache(s)
	r.Warnings = append(r.Warnings, w...)
	r.Images, w = Images(s)
	r.Warnings = append(r.Warnings, w...)
	AttachImageProjects(&r.Images, s)
	r.Logs, w = Logs(s, dataRoot)
	r.Warnings = append(r.Warnings, w...)
	r.Projects, w = Projects(s, r.Logs, now)
	r.Warnings = append(r.Warnings, w...)
	r.Volumes, w = Volumes(s)
	r.Warnings = append(r.Warnings, w...)
	r.Summary = Summarize(DfTotals(s), r.BuildCache.SharedCountedInDfBytes, r.Logs)
	r.Summary.VolumeSizesUnknown = r.Volumes.UnknownSizes
	r.Summary.Complete = r.Summary.LogBytesComplete && r.Summary.VolumeSizesUnknown == 0 && !r.BuildCache.DfRowMismatch
	return r
}
