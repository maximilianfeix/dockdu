package collect

import (
	"testing"

	"github.com/maximilianfeix/dockdu/internal/engine"
)

func TestImagesDanglingAndUniqueRules(t *testing.T) {
	s := &engine.Snapshot{
		Totals: engine.Totals{Images: engine.Aggregate{TotalCount: 5, TotalSize: 1500, Reclaimable: 700}},
		Images: []engine.Image{
			{ID: "sha256:a", RepoTags: []string{"app:1"}, Size: 500, SharedSize: 100, Containers: 1},       // in use
			{ID: "sha256:b", RepoTags: nil, Size: 300, SharedSize: 100, Containers: 0},                     // dangling, unique 200
			{ID: "sha256:c", RepoTags: []string{"<none>:<none>"}, Size: 200, SharedSize: 0, Containers: 0}, // dangling, unique 200
			{ID: "sha256:d", RepoTags: []string{"tool:2"}, Size: 300, SharedSize: -1, Containers: 0},       // unused, shared unknown
			{ID: "sha256:e", RepoTags: []string{"old:1"}, Size: 200, SharedSize: 50, Containers: -1},       // containers unknown
		},
	}
	got, warns := Images(s)
	if len(warns) != 0 {
		t.Fatalf("warnings: %v", warns)
	}
	if got.Count != 5 || got.TotalBytes != 1500 || got.ReclaimableBytes != 700 {
		t.Fatalf("count/total/reclaimable = %d/%d/%d", got.Count, got.TotalBytes, got.ReclaimableBytes)
	}
	if got.Dangling != 2 || got.DanglingBytes != 400 {
		t.Fatalf("dangling = %d (%d bytes), want 2 (400)", got.Dangling, got.DanglingBytes)
	}
	byID := map[string]int64{}
	for _, it := range got.Items {
		byID[it.ID] = it.UniqueBytes
	}
	if byID["sha256:a"] != -1 {
		t.Errorf("image in use must not claim unique bytes, got %d", byID["sha256:a"])
	}
	if byID["sha256:b"] != 200 || byID["sha256:c"] != 200 {
		t.Errorf("unique = Size-SharedSize for unused images, got b=%d c=%d", byID["sha256:b"], byID["sha256:c"])
	}
	if byID["sha256:d"] != -1 {
		t.Errorf("SharedSize -1 means unknown, got %d", byID["sha256:d"])
	}
	if byID["sha256:e"] != -1 {
		t.Errorf("Containers -1 means unknown, got %d", byID["sha256:e"])
	}
}

func TestImagesSuggestionsFollowDangling(t *testing.T) {
	s := &engine.Snapshot{Images: []engine.Image{{ID: "sha256:x", Size: 10, SharedSize: 0, Containers: 0}}}
	got, _ := Images(s)
	if len(got.Suggestions) == 0 {
		t.Fatalf("dangling image should suggest `docker image prune`")
	}
	clean := &engine.Snapshot{Images: []engine.Image{{ID: "sha256:y", RepoTags: []string{"a:1"}, Size: 10, Containers: 1}}}
	got, _ = Images(clean)
	if len(got.Suggestions) != 0 {
		t.Fatalf("nothing to prune: no suggestions, got %v", got.Suggestions)
	}
}

func TestImagesEmptyAndSorted(t *testing.T) {
	got, warns := Images(&engine.Snapshot{})
	if got.Count != 0 || len(got.Items) != 0 || got.Items == nil || len(warns) != 0 {
		t.Fatalf("empty: %+v %v", got, warns)
	}
	s := &engine.Snapshot{Images: []engine.Image{
		{ID: "sha256:small", Size: 1, Containers: 0}, {ID: "sha256:big", Size: 9, Containers: 0}, {ID: "sha256:mid", Size: 5, Containers: 0},
	}}
	got, _ = Images(s)
	if got.Items[0].ID != "sha256:big" || got.Items[2].ID != "sha256:small" {
		t.Fatalf("items must be sorted by size desc, got %v", got.Items)
	}
}

func TestImagesDuplicateIDsAndSortTieBreak(t *testing.T) {
	s := &engine.Snapshot{Images: []engine.Image{
		{ID: "sha256:b", Size: 5, Containers: 0}, {ID: "sha256:a", Size: 5, Containers: 0}, {ID: "sha256:a", Size: 5, Containers: 0},
		{ID: "sha256:d", Size: 7, SharedSize: -1, Containers: 0}, // dangling, unique unknown
	}}
	got, _ := Images(s)
	if len(got.Items) != 3 {
		t.Fatalf("duplicate image IDs must count once, got %d items", len(got.Items))
	}
	if got.Items[1].ID != "sha256:a" || got.Items[2].ID != "sha256:b" {
		t.Fatalf("equal sizes sort by ID: %v %v", got.Items[1].ID, got.Items[2].ID)
	}
	if got.DanglingUnknown != 1 || got.DanglingBytes != 5+5 {
		t.Fatalf("dangling with unknown unique size must be counted separately: unknown=%d bytes=%d", got.DanglingUnknown, got.DanglingBytes)
	}
}
