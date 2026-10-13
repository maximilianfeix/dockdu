package collect

import (
	"testing"
	"time"

	"github.com/maximilianfeix/dockdu/internal/engine"
	"github.com/maximilianfeix/dockdu/internal/report"
)

const projLabel = "com.docker.compose.project"

func composeSnapshot() *engine.Snapshot {
	now := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC).Unix()
	return &engine.Snapshot{
		Images: []engine.Image{
			{ID: "img-shop", RepoTags: []string{"shop:1"}, Size: 1000, SharedSize: 0, Containers: 2},
			{ID: "img-shared", RepoTags: []string{"postgres:16"}, Size: 500, SharedSize: 0, Containers: 2},
			{ID: "img-blog", RepoTags: []string{"blog:1"}, Size: 300, SharedSize: 0, Containers: 1},
			{ID: "img-dangling", RepoTags: nil, Size: 200, SharedSize: 0, Containers: 0},
			{ID: "img-unused", RepoTags: []string{"old:1"}, Size: 100, SharedSize: 0, Containers: 0},
		},
		Containers: []engine.Container{
			{ID: "c1", Name: "shop-web-1", ImageID: "img-shop", State: "running", Created: now, Labels: map[string]string{projLabel: "shop"}},
			{ID: "c2", Name: "shop-web-2", ImageID: "img-shop", State: "running", Created: now, Labels: map[string]string{projLabel: "shop"}},
			{ID: "c3", Name: "shop-db-1", ImageID: "img-shared", State: "running", Created: now, Labels: map[string]string{projLabel: "shop"}},
			{ID: "c4", Name: "blog-db-1", ImageID: "img-shared", State: "running", Created: now, Labels: map[string]string{projLabel: "blog"}},
			{ID: "c5", Name: "blog-app-1", ImageID: "img-blog", State: "exited", Created: now - 10*86400, Labels: map[string]string{projLabel: "blog"}},
			{ID: "c6", Name: "loose", ImageID: "img-unused", State: "exited", Created: now - 1*86400},
		},
		Volumes: []engine.Volume{
			{Name: "shop_db", HasUsageData: true, Size: 4000, RefCount: 1, Labels: map[string]string{projLabel: "shop"}},
			{Name: "shop_cache", HasUsageData: true, Size: -1, RefCount: 0, Labels: map[string]string{projLabel: "shop"}},
			{Name: "stray", HasUsageData: true, Size: 700, RefCount: 0},
			{Name: "nfs_data", Driver: "nfs", HasUsageData: false, Size: -1, RefCount: -1},
		},
	}
}

func TestProjectsGroupAndCountSharedImagesOnce(t *testing.T) {
	s := composeSnapshot()
	logs := report.Logs{Complete: true, Items: []report.LogItem{
		{ContainerID: "c1", Name: "shop-web-1", Project: "shop", Measured: true, Bytes: 50},
		{ContainerID: "c4", Name: "blog-db-1", Project: "blog", Measured: true, Bytes: 20},
		{ContainerID: "c6", Name: "loose", Measured: false},
	}}
	got, warns := Projects(s, logs, time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC))
	if len(warns) != 0 {
		t.Fatalf("warnings: %v", warns)
	}
	byName := map[string]report.Project{}
	for _, p := range got.Items {
		byName[p.Name] = p
	}
	shop, blog, none := byName["shop"], byName["blog"], byName["(none)"]
	if shop.Containers != 3 || blog.Containers != 2 || none.Containers != 1 {
		t.Fatalf("containers: shop=%d blog=%d none=%d", shop.Containers, blog.Containers, none.Containers)
	}
	// img-shop (1000) is exclusive to shop; img-shared (500) is used by shop and blog.
	if shop.ImageBytes != 1000 || shop.SharedImageBytes != 500 || shop.ImagesShared != 1 {
		t.Fatalf("shop images: %+v", shop)
	}
	if blog.ImageBytes != 300 || blog.SharedImageBytes != 500 || blog.ImagesShared != 1 {
		t.Fatalf("blog images: %+v", blog)
	}
	if got.SharedImageBytes != 500 {
		t.Fatalf("shared image bytes counted once at the top level: %d", got.SharedImageBytes)
	}
	// (none): img-unused (100) via the loose container plus the dangling
	// image (200): dangling bytes occupy space and belong to nobody, so they
	// are counted in the (none) bucket and flagged by DanglingImages.
	if none.ImageBytes != 300 || none.DanglingImages != 1 || none.OrphanVolumes != 1 || none.VolumeBytes != 700 || none.VolumesUnknown != 1 {
		t.Fatalf("(none) bucket: %+v", none)
	}
	if shop.VolumeBytes != 4000 || shop.VolumesUnknown != 1 {
		t.Fatalf("shop volumes: %+v", shop)
	}
	if shop.LogBytes != 50 || blog.LogBytes != 20 || none.LogBytes != 0 {
		t.Fatalf("log bytes: shop=%d blog=%d none=%d", shop.LogBytes, blog.LogBytes, none.LogBytes)
	}
	if blog.StoppedOld != 1 || none.StoppedOld != 0 {
		t.Fatalf("stopped >7d: blog=%d none=%d", blog.StoppedOld, none.StoppedOld)
	}
	if shop.TotalBytes != 1000+4000+50 || blog.TotalBytes != 300+20 {
		t.Fatalf("totals exclude shared image bytes: shop=%d blog=%d", shop.TotalBytes, blog.TotalBytes)
	}
	var sum int64
	for _, p := range got.Items {
		sum += p.TotalBytes
	}
	if sum != got.TotalBytes {
		t.Fatalf("project totals must add up to the section total without double counting: %d vs %d", sum, got.TotalBytes)
	}
	if got.Items[len(got.Items)-1].Name != "(none)" {
		t.Fatalf("(none) must be listed last, got %v", got.Items)
	}
	if got.Items[0].Name != "shop" {
		t.Fatalf("projects sorted by total desc, got %s first", got.Items[0].Name)
	}
}

func TestProjectsEmpty(t *testing.T) {
	got, warns := Projects(&engine.Snapshot{}, report.Logs{Complete: true}, time.Now())
	if len(got.Items) != 0 || got.Items == nil || got.Suggestions == nil || len(warns) != 0 {
		t.Fatalf("empty: %+v %v", got, warns)
	}
}

func TestVolumesReferenceStatesAndOrphans(t *testing.T) {
	got, warns := Volumes(composeSnapshot())
	if len(warns) != 0 {
		t.Fatalf("warnings: %v", warns)
	}
	byName := map[string]report.VolumeItem{}
	for _, v := range got.Items {
		byName[v.Name] = v
	}
	if byName["shop_db"].Reference != "referenced" || byName["shop_db"].Orphan {
		t.Fatalf("shop_db: %+v", byName["shop_db"])
	}
	if byName["shop_cache"].Reference != "unreferenced" || byName["shop_cache"].Orphan {
		t.Fatalf("unreferenced but labelled by compose is not an orphan: %+v", byName["shop_cache"])
	}
	if byName["stray"].Reference != "unreferenced" || !byName["stray"].Orphan {
		t.Fatalf("stray: %+v", byName["stray"])
	}
	if byName["nfs_data"].Reference != "unknown" || byName["nfs_data"].Orphan || byName["nfs_data"].Bytes != -1 {
		t.Fatalf("unknown usage must stay unknown, never orphan: %+v", byName["nfs_data"])
	}
	if got.Count != 4 || got.TotalBytes != 4700 || got.UnknownSizes != 2 {
		t.Fatalf("count/total/unknown = %d/%d/%d", got.Count, got.TotalBytes, got.UnknownSizes)
	}
	if len(got.Suggestions) != 1 || got.Suggestions[0] != report.SuggestVolumeRm("stray") {
		t.Fatalf("exactly one per-orphan suggestion, got %v", got.Suggestions)
	}
	for _, sg := range got.Suggestions {
		if containsVolumesFlag(sg) {
			t.Fatalf("bulk volume removal must never be suggested: %q", sg)
		}
	}
}

func TestAttachImageProjects(t *testing.T) {
	s := composeSnapshot()
	imgs, _ := Images(s)
	AttachImageProjects(&imgs, s)
	for _, it := range imgs.Items {
		switch it.ID {
		case "img-shared":
			if len(it.Projects) != 2 || it.Projects[0] != "blog" || it.Projects[1] != "shop" {
				t.Fatalf("shared image projects sorted: %v", it.Projects)
			}
		case "img-dangling":
			if len(it.Projects) != 0 {
				t.Fatalf("dangling image has no project, got %v", it.Projects)
			}
		case "img-unused":
			if len(it.Projects) != 1 || it.Projects[0] != "(none)" {
				t.Fatalf("image used only by a non-compose container belongs to (none), got %v", it.Projects)
			}
		}
	}
}
