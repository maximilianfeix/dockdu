package collect

import (
	"strings"
	"testing"

	"github.com/maximilianfeix/dockdu/internal/engine"
	"github.com/maximilianfeix/dockdu/internal/report"
)

func TestBuildxNode(t *testing.T) {
	cases := map[string]string{
		"buildx_buildkit_mybuilder0_state": "mybuilder0",
		"buildx_buildkit_a_b-c1_state":     "a_b-c1",
		"buildx_buildkit__state":           "",
		"buildx_buildkit_state":            "",
		"buildx_buildkit_x_state_extra":    "",
		"mybuilder0_state":                 "",
		"shop_db":                          "",
	}
	for name, want := range cases {
		if got := buildxNode(name); got != want {
			t.Errorf("buildxNode(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestBuildxBuilderName(t *testing.T) {
	cases := map[string]string{"mybuilder0": "mybuilder", "ci20": "ci2", "node1": "", "custom": "", "0": ""}
	for node, want := range cases {
		if got := buildxBuilderName(node); got != want {
			t.Errorf("buildxBuilderName(%q) = %q, want %q", node, got, want)
		}
	}
}

func TestBuildxBuildersFromStateVolumes(t *testing.T) {
	s := &engine.Snapshot{Volumes: []engine.Volume{
		{Name: "buildx_buildkit_ci0_state", Driver: "local", HasUsageData: true, Size: 7000, RefCount: 1},
		{Name: "buildx_buildkit_old0_state", Driver: "local", HasUsageData: true, Size: 300, RefCount: 0},
		{Name: "buildx_buildkit_custom_state", Driver: "local", HasUsageData: false, RefCount: -1},
		{Name: "shop_db", Driver: "local", HasUsageData: true, Size: 50, RefCount: 1},
	}}
	vols, _ := Volumes(s)
	byName := map[string]report.VolumeItem{}
	for _, v := range vols.Items {
		byName[v.Name] = v
	}
	if byName["buildx_buildkit_ci0_state"].BuildxNode != "ci0" || byName["shop_db"].BuildxNode != "" {
		t.Fatalf("buildx node not marked on volumes: %+v", vols.Items)
	}
	if !byName["buildx_buildkit_old0_state"].Orphan {
		t.Fatalf("an unreferenced builder state volume is still an orphan volume: %+v", byName["buildx_buildkit_old0_state"])
	}

	got, sugg := BuildxBuilders(vols)
	if len(got) != 3 {
		t.Fatalf("want 3 builders, got %+v", got)
	}
	// Largest first, unknown size last, like the volume list.
	if got[0].Node != "ci0" || got[0].Builder != "ci" || got[0].Bytes != 7000 || got[0].Reference != "referenced" || got[0].Volume != "buildx_buildkit_ci0_state" {
		t.Fatalf("ci0: %+v", got[0])
	}
	if got[1].Node != "old0" || got[1].Reference != "unreferenced" {
		t.Fatalf("old0: %+v", got[1])
	}
	if got[2].Node != "custom" || got[2].Builder != "" || got[2].Bytes != -1 || got[2].Reference != "unknown" {
		t.Fatalf("custom: %+v", got[2])
	}
	// One prune hint per existing builder whose name is known; the orphaned
	// state volume already has its volume rm hint, the unknown name gets
	// the lookup hint.
	if len(sugg) != 2 || sugg[0] != report.SuggestBuildxPrune("ci") || sugg[1] != report.SuggestBuildxLs {
		t.Fatalf("suggestions: %q", sugg)
	}
	for _, sg := range sugg {
		if containsVolumesFlag(sg) || strings.Contains(sg, "volume prune") {
			t.Fatalf("bulk volume removal suggested: %q", sg)
		}
	}
}

func TestBuildAttachesBuildxBuildersToBuildCache(t *testing.T) {
	s := &engine.Snapshot{Volumes: []engine.Volume{
		{Name: "buildx_buildkit_ci0_state", Driver: "local", HasUsageData: true, Size: 7000, RefCount: 1},
	}}
	r := Build(s, "", s.TakenAt, "test")
	if len(r.BuildCache.Builders) != 1 || r.BuildCache.Builders[0].Bytes != 7000 {
		t.Fatalf("builders: %+v", r.BuildCache.Builders)
	}
	found := false
	for _, sg := range r.BuildCache.Suggestions {
		found = found || sg == report.SuggestBuildxPrune("ci")
	}
	if !found {
		t.Fatalf("build cache suggestions lack the buildx prune hint: %q", r.BuildCache.Suggestions)
	}
	empty := Build(&engine.Snapshot{}, "", s.TakenAt, "test")
	if empty.BuildCache.Builders == nil {
		t.Fatal("builders must be an empty list, never null")
	}
}
