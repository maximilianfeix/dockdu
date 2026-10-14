package collect

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"

	"github.com/maximilianfeix/dockdu/internal/report"
)

// DiskBudget bounds the --disk walk. A walk that hits a budget stops and
// reports a lower bound instead of running for an unknown time.
type DiskBudget struct {
	MaxEntries int64
}

// DefaultDiskBudget allows 50 million directory entries. Large hosts have a
// few million; the context deadline set by the caller is the main bound.
var DefaultDiskBudget = DiskBudget{MaxEntries: 50_000_000}

// fsInfo is what statfs reports for one filesystem, in bytes.
type fsInfo struct {
	size, used, avail int64
}

// Filesystem reports the filesystem that holds root and, when walk is
// true, the allocated bytes under root. It returns nil and no warning when
// root is not visible (Docker Desktop host run, remote daemon): the logs
// section already explains that case. With walk, an invisible root is a
// warning because the user asked for the measurement.
func Filesystem(ctx context.Context, root string, walk bool, budget DiskBudget) (*report.Filesystem, []string) {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		if walk {
			return nil, []string{fmt.Sprintf("disk: --disk cannot measure %s: %s", root, describeRootError(err))}
		}
		return nil, nil
	}
	st, err := statFS(root)
	if err != nil {
		return nil, []string{fmt.Sprintf("disk: statfs %s: %v", root, err)}
	}
	out := &report.Filesystem{Path: root, SizeBytes: st.size, UsedBytes: st.used, AvailableBytes: st.avail}
	if !walk {
		return out, nil
	}
	du, err := walkAllocated(ctx, root, budget)
	if err != nil {
		return out, []string{fmt.Sprintf("disk: cannot walk %s: %v", root, err)}
	}
	du.OutsideDataRootBytes = -1
	if du.Complete && st.used >= du.AllocatedBytes {
		du.OutsideDataRootBytes = st.used - du.AllocatedBytes
	}
	out.DataRoot = du
	var warns []string
	if !du.Complete {
		warns = append(warns, fmt.Sprintf("disk: walk of %s incomplete (%d unreadable director(ies)%s); allocated bytes are a lower bound",
			root, du.UnreadableDirs, budgetNote(ctx, du, budget)))
	}
	return out, warns
}

func budgetNote(ctx context.Context, du *report.DiskUsage, b DiskBudget) string {
	switch {
	case ctx.Err() != nil:
		return ", time limit reached"
	case b.MaxEntries > 0 && du.Entries >= b.MaxEntries:
		return fmt.Sprintf(", entry limit %d reached", b.MaxEntries)
	}
	return ""
}

func describeRootError(err error) string {
	switch {
	case err == nil:
		return "not a directory"
	case errors.Is(err, fs.ErrNotExist):
		return "not visible from here (on Docker Desktop it is inside the VM; run the container image)"
	case errors.Is(err, fs.ErrPermission):
		return "permission denied (run as root)"
	default:
		return err.Error()
	}
}

// sortDiskItems orders by allocated bytes descending, then name.
func sortDiskItems(items []report.DiskItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].AllocatedBytes != items[j].AllocatedBytes {
			return items[i].AllocatedBytes > items[j].AllocatedBytes
		}
		return items[i].Name < items[j].Name
	})
}
