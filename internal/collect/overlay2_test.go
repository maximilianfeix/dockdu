//go:build linux || darwin

package collect

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/maximilianfeix/dockdu/internal/report"
)

// layerTree builds a data root with an overlay2 graph driver layout: two
// image layers, one container (mount and init layer), one BuildKit
// snapshot-sized directory and one large directory without any record.
func layerTree(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "docker")
	ov := filepath.Join(root, "overlay2")
	db := filepath.Join(root, "image", "overlay2", "layerdb")
	writeFile(t, filepath.Join(ov, "imgaaaaaaaaaaaaaaaaaaaaaaa", "diff", "f"), 256<<10)
	writeFile(t, filepath.Join(ov, "imgbbbbbbbbbbbbbbbbbbbbbbb", "diff", "f"), 128<<10)
	writeFile(t, filepath.Join(ov, "ctr0000000000000000000000000000000000000000000000000000000000000", "diff", "f"), 64<<10)
	writeFile(t, filepath.Join(ov, "ctr0000000000000000000000000000000000000000000000000000000000000-init", "diff", "f"), 4<<10)
	writeFile(t, filepath.Join(ov, "leftoverleftoverleftover12", "diff", "f"), 512<<10)
	writeFile(t, filepath.Join(ov, "l", "SHORT"), 1)
	put := func(rel, content string) {
		t.Helper()
		p := filepath.Join(db, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	put("sha256/c1/cache-id", "imgaaaaaaaaaaaaaaaaaaaaaaa")
	put("sha256/c2/cache-id", "imgbbbbbbbbbbbbbbbbbbbbbbb\n")
	put("mounts/abc/mount-id", "ctr0000000000000000000000000000000000000000000000000000000000000")
	put("mounts/abc/init-id", "ctr0000000000000000000000000000000000000000000000000000000000000-init")
	return root
}

func TestOverlay2AttributesLayersFromTheLayerDB(t *testing.T) {
	root := layerTree(t)
	fs, warns := Filesystem(context.Background(), root, true, DefaultDiskBudget)
	if len(warns) != 0 || fs == nil || fs.DataRoot == nil || fs.DataRoot.Overlay2 == nil {
		t.Fatalf("fs=%+v warns=%v", fs, warns)
	}
	o := fs.DataRoot.Overlay2
	if !o.LayerDB || o.ImageLayers != 2 || o.ContainerLayers != 2 || o.Unreferenced != 1 {
		t.Fatalf("attribution: %+v", o)
	}
	if o.ImageLayerBytes < 384<<10 || o.ContainerLayerBytes < 68<<10 || o.UnreferencedBytes < 512<<10 {
		t.Fatalf("bytes: %+v", o)
	}
	if len(o.LargestUnreferenced) != 1 || o.LargestUnreferenced[0].Name != "leftoverleftoverleftover12" {
		t.Fatalf("largest: %+v", o.LargestUnreferenced)
	}
	// The per-layer walk must not change the overlay2 total.
	var ov int64
	for _, it := range fs.DataRoot.Items {
		if it.Name == "overlay2" {
			ov = it.AllocatedBytes
		}
	}
	if sum := o.ImageLayerBytes + o.ContainerLayerBytes + o.UnreferencedBytes; sum > ov || ov < 964<<10 {
		t.Fatalf("overlay2 %d, attributed %d", ov, sum)
	}
}

func TestOverlay2WithoutLayerDBAttributesNothing(t *testing.T) {
	root, _ := diskTree(t) // overlay2/ but no image/overlay2/layerdb
	fs, _ := Filesystem(context.Background(), root, true, DefaultDiskBudget)
	o := fs.DataRoot.Overlay2
	if o == nil || o.LayerDB || o.UnreferencedBytes != -1 || o.ImageLayerBytes != -1 || o.Unreferenced != 0 {
		t.Fatalf("no layer DB must mean unknown, never unreferenced: %+v", o)
	}
}

func TestOverlay2NeverFollowsSymlinkedRecords(t *testing.T) {
	root := layerTree(t)
	outside := filepath.Join(filepath.Dir(root), "planted")
	if err := os.WriteFile(outside, []byte("leftoverleftoverleftover12"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "image", "overlay2", "layerdb", "sha256", "c3", "cache-id")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	fs, _ := Filesystem(context.Background(), root, true, DefaultDiskBudget)
	if o := fs.DataRoot.Overlay2; o.Unreferenced != 1 || o.ImageLayers != 2 {
		t.Fatalf("a record pointing outside the data root was used: %+v", o)
	}
}

func TestReadIDRejectsGarbage(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{"big": string(make([]byte, 300)), "path": "../../etc", "ctl": "abc\x1b[2J"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	r, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	for _, name := range []string{"big", "path", "ctl", "missing"} {
		if id := readID(r, name); id != "" {
			t.Errorf("%s: got %q", name, id)
		}
	}
}

func TestInterpretOverlay2(t *testing.T) {
	mk := func(unref, total int64) *report.Filesystem {
		return &report.Filesystem{DataRoot: &report.DiskUsage{Complete: true,
			Items:    []report.DiskItem{{Name: "overlay2", AllocatedBytes: total}},
			Overlay2: &report.Overlay2Usage{LayerDB: true, Unreferenced: 1, UnreferencedBytes: unref, NotExplainedByBuildCacheBytes: -1}}}
	}
	graph := &report.Report{Daemon: report.Daemon{ImageStore: "graphdriver"}, BuildCache: report.BuildCache{PrivateBytes: 10 << 20}}

	f := mk(5<<30, 8<<30)
	InterpretOverlay2(f, graph)
	if o := f.DataRoot.Overlay2; o.NotExplainedByBuildCacheBytes != 5<<30-10<<20 || !o.LikelyLeftovers {
		t.Fatalf("5 GiB without record: %+v", o)
	}
	// Block overhead of a few snapshot directories is not a finding.
	f = mk(28<<10, 28<<20)
	InterpretOverlay2(f, graph)
	if o := f.DataRoot.Overlay2; o.NotExplainedByBuildCacheBytes != 0 || o.LikelyLeftovers {
		t.Fatalf("small: %+v", o)
	}
	f = mk(100<<20, 100<<30) // above 64 MiB but under a tenth of overlay2
	InterpretOverlay2(f, graph)
	if f.DataRoot.Overlay2.LikelyLeftovers {
		t.Fatalf("under 10%%: %+v", f.DataRoot.Overlay2)
	}
	// Private build cache unknown: no comparison.
	f = mk(5<<30, 8<<30)
	InterpretOverlay2(f, &report.Report{Daemon: report.Daemon{ImageStore: "graphdriver"}, BuildCache: report.BuildCache{PrivateBytes: -1}})
	if o := f.DataRoot.Overlay2; o.NotExplainedByBuildCacheBytes != -1 || o.LikelyLeftovers {
		t.Fatalf("unknown private: %+v", o)
	}
	// Incomplete walk: no comparison.
	f = mk(5<<30, 8<<30)
	f.DataRoot.Complete = false
	InterpretOverlay2(f, graph)
	if f.DataRoot.Overlay2.LikelyLeftovers {
		t.Fatal("an incomplete walk must not produce a finding")
	}
	// containerd image store: overlay2/ is the hidden classic store.
	f = mk(5<<30, 8<<30)
	InterpretOverlay2(f, &report.Report{Daemon: report.Daemon{ImageStore: "containerd"}})
	if o := f.DataRoot.Overlay2; !o.HiddenByContainerdStore || o.LikelyLeftovers || o.NotExplainedByBuildCacheBytes != -1 {
		t.Fatalf("containerd: %+v", o)
	}
	InterpretOverlay2(nil, graph) // no panic
}

func TestClassicStoreLayersCountsRecordsOnly(t *testing.T) {
	if n := ClassicStoreLayers(layerTree(t)); n != 2 {
		t.Fatalf("got %d, want 2", n)
	}
	if n := ClassicStoreLayers(filepath.Join(t.TempDir(), "missing")); n != 0 {
		t.Fatalf("missing root: %d", n)
	}
}
