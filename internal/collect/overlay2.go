package collect

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/maximilianfeix/dockdu/internal/report"
)

// layerDB is where the overlay2 graph driver records which directory under
// overlay2/ belongs to which image layer or container (moby
// layer/filestore.go: cache-id per layer, mount-id and init-id per
// container).
const layerDB = "image/overlay2/layerdb"

// layerDirID is the form of a directory name under overlay2/: 26 random
// characters or a 64-hex ID, with "-init" for a container's init layer.
var layerDirID = regexp.MustCompile(`^[A-Za-z0-9]{1,128}(-init)?$`)

// maxLargestUnreferenced bounds Overlay2Usage.LargestUnreferenced.
const maxLargestUnreferenced = 10

// attributeOverlay2 splits the walked overlay2/ children into image
// layers, container layers and unreferenced directories. It opens only
// the small ID files under the layer DB, through os.Root so a symbolic
// link cannot lead outside the data root.
func attributeOverlay2(root string, layers map[string]int64) *report.Overlay2Usage {
	out := &report.Overlay2Usage{LargestUnreferenced: []report.DiskItem{}, NotExplainedByBuildCacheBytes: -1}
	images, containers, ok := layerRefs(root)
	if !ok {
		out.ImageLayerBytes, out.ContainerLayerBytes, out.UnreferencedBytes = -1, -1, -1
		return out
	}
	out.LayerDB = true
	for name, bytes := range layers {
		switch {
		case name == "l":
			// short-name symlinks for mount options, not a layer
		case images[name]:
			out.ImageLayers++
			out.ImageLayerBytes += bytes
		case containers[name]:
			out.ContainerLayers++
			out.ContainerLayerBytes += bytes
		default:
			out.Unreferenced++
			out.UnreferencedBytes += bytes
			out.LargestUnreferenced = append(out.LargestUnreferenced, report.DiskItem{Name: name, AllocatedBytes: bytes})
		}
	}
	sortDiskItems(out.LargestUnreferenced)
	if len(out.LargestUnreferenced) > maxLargestUnreferenced {
		out.LargestUnreferenced = out.LargestUnreferenced[:maxLargestUnreferenced]
	}
	return out
}

// layerRefs reads the layer DB. ok is false when its layer list cannot be
// read at all. A single unreadable record is skipped; it can only make a
// directory look unreferenced, which the report words as "no record
// found", never as garbage.
func layerRefs(root string) (images, containers map[string]bool, ok bool) {
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, nil, false
	}
	defer r.Close()
	images, containers = map[string]bool{}, map[string]bool{}
	layers, err := readDirIn(r, path.Join(layerDB, "sha256"))
	if err != nil {
		return nil, nil, false
	}
	for _, e := range layers {
		if id := readID(r, path.Join(layerDB, "sha256", e.Name(), "cache-id")); id != "" {
			images[id] = true
		}
	}
	mounts, err := readDirIn(r, path.Join(layerDB, "mounts"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, nil, false
	}
	for _, e := range mounts {
		for _, f := range []string{"mount-id", "init-id"} {
			if id := readID(r, path.Join(layerDB, "mounts", e.Name(), f)); id != "" {
				containers[id] = true
			}
		}
	}
	return images, containers, true
}

// readID returns the ID stored in name, or "" when it is missing, not a
// regular file, larger than an ID or not of the form of a directory name
// under overlay2/.
func readID(r *os.Root, name string) string {
	f, err := r.Open(name)
	if err != nil {
		return ""
	}
	defer f.Close()
	if fi, err := f.Stat(); err != nil || !fi.Mode().IsRegular() {
		return ""
	}
	b, err := io.ReadAll(io.LimitReader(f, 257))
	if err != nil || len(b) > 256 {
		return ""
	}
	id := strings.TrimSpace(string(b))
	if !layerDirID.MatchString(id) {
		return ""
	}
	return id
}

// ClassicStoreLayers counts the layer records of the classic image store
// under root without walking any layer: one directory listing.
func ClassicStoreLayers(root string) int {
	r, err := os.OpenRoot(root)
	if err != nil {
		return 0
	}
	defer r.Close()
	layers, err := readDirIn(r, path.Join(layerDB, "sha256"))
	if err != nil {
		return 0
	}
	return len(layers)
}

// InterpretOverlay2 compares the overlay2 attribution with the daemon's
// view. On the containerd image store, overlay2/ is the classic store,
// hidden but kept on disk after the switch
// (docs.docker.com/engine/storage/containerd/). On the graph driver,
// unreferenced space beyond the private build cache is space no Docker
// object accounts for.
func InterpretOverlay2(f *report.Filesystem, r *report.Report) {
	if f == nil || f.DataRoot == nil || f.DataRoot.Overlay2 == nil {
		return
	}
	o := f.DataRoot.Overlay2
	if r.Daemon.ImageStore == "containerd" {
		o.HiddenByContainerdStore = true
		return
	}
	if !o.LayerDB || !f.DataRoot.Complete || r.BuildCache.PrivateBytes < 0 {
		return
	}
	n := max(o.UnreferencedBytes-r.BuildCache.PrivateBytes, 0)
	o.NotExplainedByBuildCacheBytes = n
	var total int64
	for _, it := range f.DataRoot.Items {
		if it.Name == "overlay2" {
			total = it.AllocatedBytes
		}
	}
	o.LikelyLeftovers = n >= leftoverMinBytes && n*10 >= total
}

// leftoverMinBytes keeps block overhead of small snapshot directories
// (a few KiB each) from being reported as leftovers.
const leftoverMinBytes = 64 << 20
