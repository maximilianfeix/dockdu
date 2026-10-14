package collect

import (
	"testing"
	"time"

	"github.com/maximilianfeix/dockdu/internal/engine"
	"github.com/maximilianfeix/dockdu/internal/report"
)

func TestDfTotalsSumsTheFourDfRows(t *testing.T) {
	s := &engine.Snapshot{Totals: engine.Totals{
		Images:     engine.Aggregate{TotalSize: 100},
		Containers: engine.Aggregate{TotalSize: 10},
		Volumes:    engine.Aggregate{TotalSize: 1},
		BuildCache: engine.Aggregate{TotalSize: 1000},
	}}
	got := DfTotals(s)
	if got.DfReportedBytes != 1111 {
		t.Fatalf("DfReportedBytes = %d, want 1111", got.DfReportedBytes)
	}
}

func TestSummarizeAddsLogsAndFlagsIncomplete(t *testing.T) {
	df := report.Summary{DfReportedBytes: 1000}
	got := Summarize(df, 300, report.Logs{TotalBytes: 250, Complete: false})
	// 1000 reported by df − 300 shared build cache already inside images + 250 logs
	if got.DockerBytesTotal != 950 || got.NotShownByDf != 250 || got.BuildCacheSharedBytes != 300 {
		t.Fatalf("totals wrong: %+v", got)
	}
	if got.LogBytesComplete {
		t.Fatalf("incomplete logs must propagate")
	}
	if got.Suggestions == nil {
		t.Fatalf("suggestions must serialise as [] not null")
	}
}

func TestDfTotalsFromRecordedFixture(t *testing.T) {
	s, err := engine.LoadFixture("testdata/desktop-29.2.json")
	if err != nil {
		t.Fatalf("committed fixture must load: %v", err)
	}
	got := DfTotals(s)
	// Known values of the recorded machine (docker system df on 2026-10-13).
	if got.DfImages != 3547066761 || got.DfContainers != 12288 || got.DfVolumes != 0 || got.DfBuildCache != 1546143746 {
		t.Fatalf("df rows mapped wrong: %+v", got)
	}
	if got.DfReportedBytes != 3547066761+12288+0+1546143746 {
		t.Fatalf("DfReportedBytes = %d", got.DfReportedBytes)
	}
	if len(s.Containers) == 0 || s.Containers[0].LogDriver == "" {
		t.Fatalf("fixture should carry inspected log driver, got %+v", s.Containers)
	}
}

func TestBuildMarksSummaryIncompleteWhenVolumeSizesUnknown(t *testing.T) {
	s := &engine.Snapshot{Volumes: []engine.Volume{{Name: "nfs", Driver: "nfs", Size: -1, RefCount: -1}}}
	r := Build(s, t.TempDir(), time.Unix(0, 0), "test")
	if r.Summary.Complete || r.Summary.VolumeSizesUnknown != 1 {
		t.Fatalf("an unknown volume size makes the total a lower bound: %+v", r.Summary)
	}
	r = Build(&engine.Snapshot{}, t.TempDir(), time.Unix(0, 0), "test")
	if !r.Summary.Complete {
		t.Fatalf("empty snapshot is complete: %+v", r.Summary)
	}
}
