package collect

import (
	"fmt"
	"sort"

	"github.com/maximilianfeix/dockdu/internal/engine"
	"github.com/maximilianfeix/dockdu/internal/report"
)

// BuildCache applies SPEC §6.4. Reclaimable is the daemon's number (what
// `docker system df` prints); shared, private, in-use and not-freed are
// derived from the records. Duplicate record IDs are counted once.
func BuildCache(s *engine.Snapshot) (report.BuildCache, []string) {
	var warns []string
	seen := make(map[string]bool, len(s.BuildCache))
	dups := 0
	var sum, shared, inUse, notFreed int64
	records := 0
	for _, r := range s.BuildCache {
		if r.ID != "" && seen[r.ID] {
			dups++
			continue
		}
		seen[r.ID] = true
		records++
		sum += r.Size
		if r.Shared {
			shared += r.Size
		}
		if r.InUse {
			inUse += r.Size
		}
		if r.Shared || r.InUse {
			notFreed += r.Size
		}
	}
	if dups > 0 {
		warns = append(warns, fmt.Sprintf("build cache: %d duplicate record id(s) ignored", dups))
	}

	total := s.Totals.BuildCache.TotalSize
	if total == 0 && sum > 0 {
		warns = append(warns, fmt.Sprintf("build cache: daemon total is 0 but records sum to %d bytes; showing the records sum", sum))
		total = sum
	} else if total > 0 && sum > 0 && total != sum {
		warns = append(warns, fmt.Sprintf("build cache: daemon total %d bytes differs from records sum %d bytes; showing the daemon total", total, sum))
	}

	out := report.BuildCache{
		Records:          records,
		TotalBytes:       total,
		SharedBytes:      shared,
		PrivateBytes:     total - shared,
		ReclaimableBytes: s.Totals.BuildCache.Reclaimable,
		InUseBytes:       inUse,
		NotFreedByPrune:  notFreed,
		Suggestions:      []string{},
	}
	if out.PrivateBytes < 0 {
		// The daemon total and the records disagree; unknown, not 0.
		out.PrivateBytes = -1
	}
	// Engine rule (moby builder.go): reclaimable + Σ(shared || in use) == total.
	if records > 0 && s.Totals.BuildCache.TotalSize > 0 && out.ReclaimableBytes+notFreed != total {
		warns = append(warns, fmt.Sprintf("build cache: daemon reclaimable %d + not-freed %d ≠ total %d; the reclaimable figure is the daemon's and may be stale", out.ReclaimableBytes, notFreed, total))
	}
	if out.ReclaimableBytes > 0 {
		out.Suggestions = append(out.Suggestions, report.SuggestBuilderPruneOld, report.SuggestBuilderPruneAll)
	}
	return out, warns
}

// sortCacheRecords orders records by size descending, then ID, for stable
// output. Exported collectors use it before rendering lists.
func sortCacheRecords(rs []engine.CacheRecord) {
	sort.SliceStable(rs, func(i, j int) bool {
		if rs[i].Size != rs[j].Size {
			return rs[i].Size > rs[j].Size
		}
		return rs[i].ID < rs[j].ID
	})
}
