package collect

import (
	"strings"
	"testing"
	"time"

	"github.com/maximilianfeix/dockdu/internal/engine"
)

func rec(id string, size int64, shared, inUse bool) engine.CacheRecord {
	return engine.CacheRecord{ID: id, Type: "regular", Size: size, Shared: shared, InUse: inUse, CreatedAt: time.Unix(0, 0)}
}

func TestBuildCacheUsesEngineReclaimableNotBuildxRule(t *testing.T) {
	// Observed 2026-10-13: engine reclaimable excludes shared records,
	// buildx counts every !InUse record. dockdu reports the engine number.
	s := &engine.Snapshot{
		// Engine rule: reclaimable = Σ Size(!InUse && !Shared) = record b = 200,
		// and reclaimable + Σ Size(Shared || InUse) == total (200 + 400 == 600).
		Totals: engine.Totals{BuildCache: engine.Aggregate{TotalCount: 3, TotalSize: 600, Reclaimable: 200}},
		BuildCache: []engine.CacheRecord{
			rec("a", 300, true, false),  // shared, unused
			rec("b", 200, false, false), // private, unused
			rec("c", 100, false, true),  // private, in use
		},
	}
	got, warns := BuildCache(s)
	if len(warns) != 0 {
		t.Fatalf("unexpected warnings: %v", warns)
	}
	if got.Records != 3 || got.TotalBytes != 600 {
		t.Fatalf("records/total = %d/%d", got.Records, got.TotalBytes)
	}
	if got.ReclaimableBytes != 200 {
		t.Fatalf("reclaimable = %d, want engine value 200 (not buildx 500)", got.ReclaimableBytes)
	}
	if got.SharedBytes != 300 || got.PrivateBytes != 300 || got.InUseBytes != 100 {
		t.Fatalf("shared/private/inuse = %d/%d/%d", got.SharedBytes, got.PrivateBytes, got.InUseBytes)
	}
	if got.NotFreedByPrune != 400 {
		t.Fatalf("not freed by prune = %d, want 400 (shared 300 + in use 100)", got.NotFreedByPrune)
	}
}

func TestBuildCacheDuplicateRecordsCountedOnce(t *testing.T) {
	s := &engine.Snapshot{
		Totals:     engine.Totals{BuildCache: engine.Aggregate{TotalCount: 1, TotalSize: 50, Reclaimable: 50}},
		BuildCache: []engine.CacheRecord{rec("dup", 50, false, false), rec("dup", 50, false, false)},
	}
	got, warns := BuildCache(s)
	if got.Records != 1 || got.PrivateBytes != 50 {
		t.Fatalf("duplicates must be deduplicated by ID: %+v", got)
	}
	if len(warns) != 1 {
		t.Fatalf("expected one warning about duplicates, got %v", warns)
	}
}

func TestBuildCacheEmpty(t *testing.T) {
	got, warns := BuildCache(&engine.Snapshot{})
	if got.Records != 0 || got.TotalBytes != 0 || got.NotFreedByPrune != 0 || len(warns) != 0 {
		t.Fatalf("empty cache: %+v %v", got, warns)
	}
	if got.Suggestions == nil || len(got.Suggestions) != 0 {
		t.Fatalf("empty cache must have an empty, non-nil suggestion list, got %#v", got.Suggestions)
	}
}

func TestBuildCacheTotalMismatchIsWarned(t *testing.T) {
	// API total says 0 but records sum to 10: report the records sum and say so.
	s := &engine.Snapshot{BuildCache: []engine.CacheRecord{rec("x", 10, false, false)}}
	got, warns := BuildCache(s)
	if got.TotalBytes != 10 {
		t.Fatalf("total should fall back to the records sum, got %d", got.TotalBytes)
	}
	if len(warns) != 1 {
		t.Fatalf("expected a mismatch warning, got %v", warns)
	}
}

func TestBuildCacheSuggestions(t *testing.T) {
	s := &engine.Snapshot{
		Totals:     engine.Totals{BuildCache: engine.Aggregate{TotalSize: 100, Reclaimable: 100}},
		BuildCache: []engine.CacheRecord{rec("p", 100, false, false)},
	}
	got, _ := BuildCache(s)
	if len(got.Suggestions) == 0 {
		t.Fatalf("reclaimable cache should suggest a builder prune")
	}
	for _, sg := range got.Suggestions {
		if containsVolumesFlag(sg) {
			t.Fatalf("suggestion must never touch volumes: %q", sg)
		}
	}
	none, _ := BuildCache(&engine.Snapshot{Totals: engine.Totals{BuildCache: engine.Aggregate{TotalSize: 100}},
		BuildCache: []engine.CacheRecord{rec("s", 100, true, true)}})
	if len(none.Suggestions) != 0 {
		t.Fatalf("nothing reclaimable: no prune suggestion, got %v", none.Suggestions)
	}
}

func TestBuildCacheInvariantViolationIsWarned(t *testing.T) {
	s := &engine.Snapshot{
		Totals:     engine.Totals{BuildCache: engine.Aggregate{TotalSize: 600, Reclaimable: 300}}, // should be 200
		BuildCache: []engine.CacheRecord{rec("a", 300, true, false), rec("b", 200, false, false), rec("c", 100, false, true)},
	}
	_, warns := BuildCache(s)
	if len(warns) != 1 || !strings.Contains(warns[0], "≠ total") {
		t.Fatalf("expected an invariant warning, got %v", warns)
	}
}

func TestBuildCacheNegativePrivateIsUnknown(t *testing.T) {
	s := &engine.Snapshot{
		Totals:     engine.Totals{BuildCache: engine.Aggregate{TotalSize: 100, Reclaimable: 0}},
		BuildCache: []engine.CacheRecord{rec("a", 300, true, false)},
	}
	got, _ := BuildCache(s)
	if got.PrivateBytes != -1 {
		t.Fatalf("private must be unknown (-1) when records exceed the daemon total, got %d", got.PrivateBytes)
	}
}

// legacyCache is what moby/client v0.6.2 synthesises for daemons on API
// < 1.52 (Docker Engine 28 and older): TotalSize and Reclaimable are
// computed client-side from the records and leave shared records out.
func legacyCache(api string) *engine.Snapshot {
	return &engine.Snapshot{
		Daemon: engine.Daemon{APIVersion: api},
		// TotalSize = Σ !Shared = 200 + 100; Reclaimable = Σ !Shared && !InUse = 200.
		Totals: engine.Totals{BuildCache: engine.Aggregate{TotalCount: 3, TotalSize: 300, Reclaimable: 200}},
		BuildCache: []engine.CacheRecord{
			rec("a", 300, true, false),
			rec("b", 200, false, false),
			rec("c", 100, false, true),
		},
	}
}

func TestBuildCacheLegacyAPITotalExcludesShared(t *testing.T) {
	for _, api := range []string{"1.51", "1.44", ""} {
		got, warns := BuildCache(legacyCache(api))
		if len(warns) != 0 {
			t.Fatalf("api %q: a legacy total is expected, not a mismatch: %v", api, warns)
		}
		if got.TotalBytes != 600 || got.SharedBytes != 300 || got.PrivateBytes != 300 {
			t.Fatalf("api %q: total/shared/private = %d/%d/%d, want 600/300/300", api, got.TotalBytes, got.SharedBytes, got.PrivateBytes)
		}
		if got.SharedCountedInDfBytes != 0 {
			t.Fatalf("api %q: df's build cache row has no shared bytes to subtract, got %d", api, got.SharedCountedInDfBytes)
		}
		if got.ReclaimableBytes != 200 || got.NotFreedByPrune != 400 {
			t.Fatalf("api %q: reclaimable/not freed = %d/%d", api, got.ReclaimableBytes, got.NotFreedByPrune)
		}
	}
}

func TestBuildCacheCurrentAPICountsSharedInDf(t *testing.T) {
	s := legacyCache("1.53")
	s.Totals.BuildCache.TotalSize = 600 // API >= 1.52: the daemon total includes shared records
	got, warns := BuildCache(s)
	if len(warns) != 0 || got.TotalBytes != 600 || got.SharedCountedInDfBytes != 300 {
		t.Fatalf("got %+v, warnings %v", got, warns)
	}
}

func TestBuildCacheSharedSubtractionNeverExceedsDfRow(t *testing.T) {
	// The daemon reports 0 for the row but records exist: nothing of the row
	// can be subtracted, whatever the records say.
	s := &engine.Snapshot{Daemon: engine.Daemon{APIVersion: "1.53"}, BuildCache: []engine.CacheRecord{rec("a", 50, true, false)}}
	got, _ := BuildCache(s)
	if got.SharedCountedInDfBytes != 0 {
		t.Fatalf("shared subtracted from df must be ≤ the df row (0), got %d", got.SharedCountedInDfBytes)
	}
}

func TestBuildLegacyDaemonSummaryDoesNotSubtractSharedTwice(t *testing.T) {
	s := legacyCache("1.51")
	s.Totals.Images.TotalSize = 1000
	r := Build(s, t.TempDir(), time.Unix(0, 0), "test")
	// df = images 1000 + build cache 300 (already without shared); no logs.
	if r.Summary.DfReportedBytes != 1300 || r.Summary.DockerBytesTotal != 1300 || r.Summary.BuildCacheSharedBytes != 0 {
		t.Fatalf("summary %+v", r.Summary)
	}
}
