package collect

import (
	"fmt"
	"strconv"
	"strings"

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

	// Which build-cache total df shows depends on the API version: daemons
	// on API >= 1.52 send a TotalSize that includes shared records; for older
	// daemons moby/client computes TotalSize from the records and leaves the
	// shared ones out (client system_disk_usage.go, legacy path). The
	// numbers decide first, the version only when they match neither rule.
	dfTotal := s.Totals.BuildCache.TotalSize
	private := sum - shared
	total, sharedInDf := dfTotal, shared
	legacy := apiBefore(s.Daemon.APIVersion, 1, 52)
	excludesShared, mismatch := false, false
	switch {
	case records == 0 || dfTotal == sum:
	// A row of 0 with only shared records is the legacy rule only on an old
	// daemon; on a current one it is a mismatch.
	case dfTotal == private && shared > 0 && (dfTotal > 0 || legacy):
		total, sharedInDf, excludesShared = sum, 0, true
	case dfTotal == 0:
		warns = append(warns, fmt.Sprintf("build cache: daemon total is 0 but records sum to %d bytes; showing the records sum, the summary total is a lower bound", sum))
		total, sharedInDf, mismatch = sum, 0, true
	case legacy:
		warns = append(warns, fmt.Sprintf("build cache: daemon total %d bytes matches neither the records sum %d nor the unshared sum %d; showing the records sum, the summary total is a lower bound", dfTotal, sum, private))
		total, sharedInDf, mismatch = sum, 0, true
	default:
		warns = append(warns, fmt.Sprintf("build cache: daemon total %d bytes differs from records sum %d bytes; showing the daemon total, the summary total is a lower bound", dfTotal, sum))
		mismatch = true
	}
	// Never subtract more than the df row holds.
	sharedInDf = min(sharedInDf, max(dfTotal, 0))

	out := report.BuildCache{
		Records:                records,
		TotalBytes:             total,
		SharedBytes:            shared,
		PrivateBytes:           total - shared,
		ReclaimableBytes:       s.Totals.BuildCache.Reclaimable,
		InUseBytes:             inUse,
		NotFreedByPrune:        notFreed,
		SharedCountedInDfBytes: sharedInDf,
		DfRowExcludesShared:    excludesShared,
		DfRowMismatch:          mismatch,
		Builders:               []report.BuildxBuilder{},
		Suggestions:            []string{},
	}
	if containerdStore(s) && shared > 0 {
		// A shared record can keep its own copy on the containerd image
		// store; how much a prune frees is not reported by Docker.
		out.SharedMayBeFreedByPrune = true
		out.NotFreedByPrune = -1
	}
	if out.PrivateBytes < 0 {
		// The daemon total and the records disagree; unknown, not 0.
		out.PrivateBytes = -1
	}
	// Engine rule (moby builder.go): reclaimable + Σ(shared || in use) == Σ
	// records. It holds for the legacy client computation too.
	if records > 0 && dfTotal > 0 && out.ReclaimableBytes+notFreed != sum && !out.SharedMayBeFreedByPrune {
		warns = append(warns, fmt.Sprintf("build cache: daemon reclaimable %d + not-freed %d ≠ total %d; the reclaimable figure is the daemon's and may be stale", out.ReclaimableBytes, notFreed, sum))
	}
	if out.ReclaimableBytes > 0 {
		out.Suggestions = append(out.Suggestions, report.SuggestBuilderPruneOld, report.SuggestBuilderPruneAll)
	}
	return out, warns
}

// apiBefore reports whether the negotiated API version ("1.51") is older
// than major.minor. An unparsable or empty version is not "before".
func apiBefore(v string, major, minor int) bool {
	maj, mn, ok := strings.Cut(v, ".")
	if !ok {
		return false
	}
	a, err1 := strconv.Atoi(maj)
	b, err2 := strconv.Atoi(mn)
	if err1 != nil || err2 != nil {
		return false
	}
	return a < major || (a == major && b < minor)
}
