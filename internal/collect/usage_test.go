package collect

import (
	"testing"

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
	got := Summarize(df, report.Logs{TotalBytes: 250, Complete: false})
	if got.DockerBytesTotal != 1250 || got.NotShownByDf != 250 {
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
		t.Skipf("fixture not recorded yet: %v", err)
	}
	got := DfTotals(s)
	want := s.Totals.Images.TotalSize + s.Totals.Containers.TotalSize + s.Totals.Volumes.TotalSize + s.Totals.BuildCache.TotalSize
	if got.DfReportedBytes != want || want <= 0 {
		t.Fatalf("DfReportedBytes = %d, want %d (>0)", got.DfReportedBytes, want)
	}
	if len(s.Containers) == 0 || s.Containers[0].LogDriver == "" {
		t.Fatalf("fixture should carry inspected log driver, got %+v", s.Containers)
	}
}
