//go:build linux || darwin

package collect

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// diskTree builds a small data root: overlay2 with a 64 KiB file and a hard
// link to it, containers with a log, a symlink to a large file outside the
// root, and a file directly in the root.
func diskTree(t *testing.T) (root, outside string) {
	t.Helper()
	base := t.TempDir()
	root = filepath.Join(base, "docker")
	outside = filepath.Join(base, "outside.bin")
	writeFile(t, outside, 4<<20) // 4 MiB, must never be counted
	writeFile(t, filepath.Join(root, "overlay2", "l1", "diff", "blob"), 64<<10)
	if err := os.Link(filepath.Join(root, "overlay2", "l1", "diff", "blob"), filepath.Join(root, "overlay2", "l1", "diff", "blob-link")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "containers", "c1", "c1-json.log"), 32<<10)
	if err := os.Symlink(outside, filepath.Join(root, "containers", "c1", "evil")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "engine-id"), 10)
	return root, outside
}

func TestFilesystemWithoutWalkReportsStatfsOnly(t *testing.T) {
	root, _ := diskTree(t)
	fs, warns := Filesystem(context.Background(), root, false, DefaultDiskBudget)
	if len(warns) != 0 || fs == nil {
		t.Fatalf("fs=%+v warns=%v", fs, warns)
	}
	if fs.SizeBytes <= 0 || fs.UsedBytes <= 0 || fs.UsedBytes > fs.SizeBytes || fs.AvailableBytes > fs.SizeBytes {
		t.Fatalf("implausible statfs numbers: %+v", fs)
	}
	if fs.DataRoot != nil {
		t.Fatal("no walk without --disk")
	}
}

func TestFilesystemWalkCountsInodesOnceAndNeverFollowsSymlinks(t *testing.T) {
	root, _ := diskTree(t)
	fs, warns := Filesystem(context.Background(), root, true, DefaultDiskBudget)
	if len(warns) != 0 || fs == nil || fs.DataRoot == nil {
		t.Fatalf("fs=%+v warns=%v", fs, warns)
	}
	du := fs.DataRoot
	if !du.Complete || du.UnreadableDirs != 0 {
		t.Fatalf("walk should be complete: %+v", du)
	}
	byName := map[string]int64{}
	for _, it := range du.Items {
		byName[it.Name] = it.AllocatedBytes
	}
	// The 64 KiB blob and its hard link occupy one inode: at least 64 KiB,
	// well under 128 KiB (plus directory blocks).
	if ov := byName["overlay2"]; ov < 64<<10 || ov >= 128<<10 {
		t.Fatalf("overlay2 = %d, want one copy of the 64 KiB blob", ov)
	}
	// The symlink target (4 MiB) is outside the root and must not count.
	if c := byName["containers"]; c >= 1<<20 {
		t.Fatalf("containers = %d: the symlink was followed", c)
	}
	if _, ok := byName[rootFilesItem]; !ok {
		t.Fatalf("files directly in the root are grouped as %q: %+v", rootFilesItem, du.Items)
	}
	var sum int64
	for _, it := range du.Items {
		sum += it.AllocatedBytes
	}
	if sum > du.AllocatedBytes || du.AllocatedBytes >= 1<<20 {
		t.Fatalf("items %d, total %d", sum, du.AllocatedBytes)
	}
	// Both are block counts of one filesystem, so the difference is defined.
	if du.OutsideDataRootBytes != fs.UsedBytes-du.AllocatedBytes {
		t.Fatalf("outside = %d, want used %d − allocated %d", du.OutsideDataRootBytes, fs.UsedBytes, du.AllocatedBytes)
	}
	if du.Items[0].AllocatedBytes < du.Items[len(du.Items)-1].AllocatedBytes {
		t.Fatalf("items not sorted largest first: %+v", du.Items)
	}
}

func TestFilesystemUnreadableDirMakesALowerBound(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads the directory anyway")
	}
	root, _ := diskTree(t)
	locked := filepath.Join(root, "volumes", "secret")
	writeFile(t, filepath.Join(locked, "data"), 8<<10)
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })
	fs, warns := Filesystem(context.Background(), root, true, DefaultDiskBudget)
	du := fs.DataRoot
	if du.Complete || du.UnreadableDirs != 1 || du.OutsideDataRootBytes != -1 {
		t.Fatalf("an unreadable directory must make the walk incomplete with unknown outside bytes: %+v", du)
	}
	if len(warns) != 1 || !strings.Contains(warns[0], "lower bound") {
		t.Fatalf("want one lower-bound warning, got %v", warns)
	}
}

func TestFilesystemBudgetsStopTheWalk(t *testing.T) {
	root, _ := diskTree(t)
	fs, warns := Filesystem(context.Background(), root, true, DiskBudget{MaxEntries: 3})
	if fs.DataRoot.Complete || fs.DataRoot.Entries > 3 || len(warns) != 1 || !strings.Contains(warns[0], "entry limit 3") {
		t.Fatalf("entry budget not enforced: %+v %v", fs.DataRoot, warns)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fs, warns = Filesystem(ctx, root, true, DefaultDiskBudget)
	if fs.DataRoot.Complete || len(warns) != 1 || !strings.Contains(warns[0], "time limit") {
		t.Fatalf("a cancelled walk must be incomplete: %+v %v", fs.DataRoot, warns)
	}
}

func TestFilesystemInvisibleRoot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "inside-the-vm")
	if fs, warns := Filesystem(context.Background(), missing, false, DefaultDiskBudget); fs != nil || len(warns) != 0 {
		t.Fatalf("without --disk an invisible root is silent: %+v %v", fs, warns)
	}
	fs, warns := Filesystem(context.Background(), missing, true, DefaultDiskBudget)
	if fs != nil || len(warns) != 1 || !strings.Contains(warns[0], "not visible from here") {
		t.Fatalf("with --disk an invisible root is a warning: %+v %v", fs, warns)
	}
}
