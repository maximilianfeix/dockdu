package report

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "rewrite golden files")

// goldenReport is a hand-built report that exercises every section,
// including unknown values, incomplete logs, shared images and warnings.
func goldenReport() *Report {
	return &Report{
		Version: "v0.1.0-test",
		TakenAt: time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC),
		Daemon:  Daemon{ServerVersion: "29.2.0", APIVersion: "1.53", OperatingSystem: "Docker Desktop", DataRoot: "/var/lib/docker"},
		Summary: Summary{DfReportedBytes: 90403946496, DfImages: 5368709120, DfContainers: 104857600, DfVolumes: 16106127360, DfBuildCache: 68824252416,
			LogBytes: 51861225472, LogBytesComplete: false, VolumeSizesUnknown: 1, BuildCacheSharedBytes: 33285996544, DockerBytesTotal: 108979175424, NotShownByDf: 51861225472, Suggestions: []string{}},
		BuildCache: BuildCache{Records: 26, TotalBytes: 68824252416, SharedBytes: 33285996544, PrivateBytes: 35538255872, ReclaimableBytes: 35538255872, NotFreedByPrune: 33285996544,
			Suggestions: []string{SuggestBuilderPruneOld}},
		Logs: Logs{TotalBytes: 51861225472, Complete: false, Items: []LogItem{
			{ContainerID: "a", Name: "shop-web-1", Project: "shop", Driver: "json-file", Rotation: "unbounded", Bytes: 6657199104, Files: 1, Measured: true},
			{ContainerID: "b", Name: "blog-worker-3", Project: "blog", Driver: "local", Rotation: "bounded", MaxSize: "20m (driver default)", MaxFile: "5 (driver default)", Bytes: 104857600, Files: 5, Measured: true},
			{ContainerID: "c", Name: "loose", Driver: "json-file", Rotation: "unbounded", Bytes: -1, Measured: false, Note: "log directory not readable (run as root or inside the container image)"},
		}, Suggestions: []string{SuggestLogOptsDaemon}},
		// TotalBytes = Σ items (23837068287 + 15073741824) + shared 500000000.
		Projects: Projects{TotalBytes: 39410810111, SharedImageBytes: 500000000, ImageOverlapBytes: 300000000, Items: []Project{
			{Name: "shop", Containers: 4, ImageBytes: 4187593113, SharedImageBytes: 500000000, ImagesShared: 1, VolumeBytes: 12992276070, LogBytes: 6657199104, TotalBytes: 23837068287},
			{Name: "(none)", Containers: 2, ImageBytes: 1073741824, VolumeBytes: 14000000000, VolumesUnknown: 1, LogBytes: 0, LogsUnknown: 1, TotalBytes: 15073741824, OrphanVolumes: 3, DanglingImages: 7, StoppedOld: 1},
		}, Suggestions: []string{}},
		Images: Images{Count: 3, TotalBytes: 5368709120, Dangling: 1, DanglingBytes: 209715200, ReclaimableBytes: 1181116006, Items: []ImageItem{
			{ID: "sha256:aaaaaaaaaaaaaaaa", Tags: []string{"shop:1", "shop:latest"}, Bytes: 4187593113, SharedBytes: 0, Containers: 2, UniqueBytes: -1, Projects: []string{"shop"}},
			{ID: "sha256:bbbbbbbbbbbbbbbb", Tags: []string{"postgres:16"}, Bytes: 500000000, SharedBytes: 100000000, Containers: 2, UniqueBytes: -1, Projects: []string{"blog", "shop"}},
			{ID: "sha256:cccccccccccccccc", Tags: nil, Bytes: 209715200, SharedBytes: 0, Containers: -1, Dangling: true, UniqueBytes: -1},
		}, Suggestions: []string{SuggestImagePruneDangling}},
		Volumes: Volumes{Count: 3, TotalBytes: 26992276070, UnknownSizes: 1, Items: []VolumeItem{
			{Name: "shop_db", Driver: "local", Project: "shop", Bytes: 12992276070, RefCount: 1, Reference: "referenced"},
			{Name: "stray", Driver: "local", Bytes: 14000000000, RefCount: 0, Reference: "unreferenced", Orphan: true},
			{Name: "nfs_data", Driver: "nfs", Bytes: -1, RefCount: -1, Reference: "unknown"},
		}, Suggestions: []string{SuggestVolumeRm("stray")}},
		Desktop: &Desktop{Product: "Docker Desktop", DiskFile: "/Users/<user>/Library/Containers/com.docker.docker/Data/vms/0/data/Docker.raw",
			ApparentBytes: 245106737152, AllocatedBytes: 4779278336, DockerReportedBytes: 108979175424, Explanation: "Explanation paragraph."},
		Warnings: []string{"logs: loose not measured: log directory not readable (run as root or inside the container image)"},
	}
}

func TestWriteTextGolden(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteText(&buf, goldenReport(), 10); err != nil {
		t.Fatal(err)
	}
	const path = "testdata/report.golden"
	if *update {
		if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden file missing (run with -update): %v", err)
	}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("text output differs from golden file; run `go test ./internal/report -update` after checking the diff:\n%s", buf.String())
	}
}

func TestWriteTextTopLimitsRowsAndSaysSo(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteText(&buf, goldenReport(), 1); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "CONTAINER LOGS  (top 1 of 3") || strings.Contains(out, "blog-worker-3") {
		t.Fatalf("--top 1 must show one log row and say 'top 1 of 3':\n%s", out)
	}
	if !strings.Contains(flat(out), "IMAGES top 1 of 3 · total") {
		t.Fatalf("images header must report the limit:\n%s", out)
	}
}

func TestWriteTextNeverSuggestsBulkVolumeRemoval(t *testing.T) {
	var buf bytes.Buffer
	_ = WriteText(&buf, goldenReport(), 0)
	if strings.Contains(buf.String(), "volume prune") || strings.Contains(buf.String(), "--volumes") {
		t.Fatalf("bulk volume removal leaked into the report")
	}
}

func TestWriteJSONListsAreNeverNull(t *testing.T) {
	r := &Report{Version: "x"}
	var buf bytes.Buffer
	if err := WriteJSON(&buf, r); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), `"warnings": null`) {
		t.Fatalf("warnings must serialise as []: %s", buf.String())
	}
}

func TestWriteTextNeutralisesControlCharacters(t *testing.T) {
	r := goldenReport()
	r.Projects.Items[0].Name = "shop\x1b[31mred"
	r.Logs.Items[0].Name = "web\x07bell"
	var buf bytes.Buffer
	_ = WriteText(&buf, r, 0)
	if strings.ContainsAny(buf.String(), "\x1b\x07") {
		t.Fatalf("control characters from the daemon reached the terminal")
	}
}

func TestWriteTextHasNoTrailingSpaces(t *testing.T) {
	var buf bytes.Buffer
	_ = WriteText(&buf, goldenReport(), 0)
	for i, l := range strings.Split(buf.String(), "\n") {
		if strings.HasSuffix(l, " ") {
			t.Fatalf("line %d has trailing spaces: %q", i+1, l)
		}
	}
}

func TestWriteTextLegacyDaemonSaysNothingWasSubtracted(t *testing.T) {
	r := goldenReport()
	r.Summary.BuildCacheSharedBytes = 0 // df row of an API < 1.52 daemon has no shared bytes
	r.BuildCache.DfRowExcludesShared = true
	var buf bytes.Buffer
	if err := WriteText(&buf, r, 10); err != nil {
		t.Fatal(err)
	}
	if out := flat(buf.String()); !strings.Contains(out, "= df + logs (df's build cache row already leaves out 31.0 GiB shared with images)") || strings.Contains(out, "= df − 0 B") {
		t.Fatalf("legacy summary line wrong:\n%s", out)
	}
}

func TestSuggestVolumeRmQuotesDaemonNames(t *testing.T) {
	cases := map[string]string{
		"pgdata":          "docker volume rm pgdata   #",
		"x; rm -rf ~":     "docker volume rm 'x; rm -rf ~'   #",
		"it's":            `docker volume rm 'it'\''s'   #`,
		"$(touch /tmp/p)": "docker volume rm '$(touch /tmp/p)'   #",
		"-rf":             "docker volume rm -- '-rf'   #",
	}
	for name, prefix := range cases {
		if got := SuggestVolumeRm(name); !strings.HasPrefix(got, prefix) {
			t.Errorf("SuggestVolumeRm(%q) = %q, want prefix %q", name, got, prefix)
		}
	}
	got := SuggestVolumeRm("evil\x1b]0;pwned\x07")
	if strings.HasPrefix(got, "docker") || strings.ContainsAny(got, "\x1b\x07") {
		t.Errorf("a name with control characters must not become a command or reach the terminal raw: %q", got)
	}
}

func TestWriteTextNeutralisesControlCharactersInSuggestions(t *testing.T) {
	r := goldenReport()
	r.Volumes.Suggestions = []string{"docker volume rm x\x1b[2J"}
	var buf bytes.Buffer
	if err := WriteText(&buf, r, 10); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "\x1b") {
		t.Fatal("escape sequence reached the terminal output")
	}
}

func TestSuggestVolumeRmRefusesBidiAndSeparators(t *testing.T) {
	for _, name := range []string{"data\u202Egnp.exe", "a\u2028b", "zero\u200Bwidth"} {
		if got := SuggestVolumeRm(name); strings.HasPrefix(got, "docker") || !strings.HasPrefix(got, "# volume") {
			t.Errorf("%q must not become a command: %q", name, got)
		}
	}
}

func TestWriteTextMarksBuildCacheMismatchAsLowerBound(t *testing.T) {
	r := goldenReport()
	r.BuildCache.DfRowMismatch = true
	var buf bytes.Buffer
	if err := WriteText(&buf, r, 10); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "build cache row disagrees with its records") {
		t.Fatalf("mismatch must be named in the lower-bound note:\n%s", buf.String())
	}
}

func TestWriteTextDataRootOnDisk(t *testing.T) {
	r := goldenReport()
	r.Filesystem = &Filesystem{Path: "/var/lib/docker", SizeBytes: 100 << 30, UsedBytes: 92 << 30, AvailableBytes: 8 << 30,
		DataRoot: &DiskUsage{AllocatedBytes: 70 << 30, Entries: 1234, Complete: true, SkippedMounts: 3, OutsideDataRootBytes: 22 << 30,
			Items: []DiskItem{{Name: "overlay2", AllocatedBytes: 50 << 30, Entries: 1000}, {Name: "containers", AllocatedBytes: 20 << 30, Entries: 234}}}}
	var buf bytes.Buffer
	if err := WriteText(&buf, r, 10); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"data root filesystem        92.0 GiB    used of 100.0 GiB · 8.0 GiB free",
		"DATA ROOT ON DISK",
		"/var/lib/docker · allocated blocks · 1234 entries · complete",
		"overlay2",
		"data root on disk           70.0 GiB",
		"outside the data root       22.0 GiB",
		"3 mount(s) inside the data root not entered",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	r.Filesystem.DataRoot.Complete, r.Filesystem.DataRoot.OutsideDataRootBytes = false, -1
	buf.Reset()
	if err := WriteText(&buf, r, 10); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "outside the data root       unknown") || !strings.Contains(buf.String(), "lower bound") {
		t.Fatalf("incomplete walk must not show an outside figure:\n%s", buf.String())
	}
}

func TestWriteTextContainerdStorePruneEffectUnknown(t *testing.T) {
	r := goldenReport()
	r.BuildCache.SharedMayBeFreedByPrune, r.BuildCache.NotFreedByPrune = true, -1
	var buf bytes.Buffer
	if err := WriteText(&buf, r, 10); err != nil {
		t.Fatal(err)
	}
	out := flat(buf.String())
	if strings.Contains(out, "prune would NOT free") || !strings.Contains(out, "prune frees an unknown part of the 31.0 GiB shared with images") ||
		!strings.Contains(out, "shared build cache may keep its own copy (containerd store)") {
		t.Fatalf("containerd store wording missing:\n%s", out)
	}
}

// flat collapses all whitespace, so assertions about content do not depend
// on where the renderer wraps lines.
func flat(s string) string { return strings.Join(strings.Fields(s), " ") }

// worstCase is a report with the widest values the renderer prints (10-rune
// sizes such as "1023.9 TiB") and every optional note switched on, but no
// list items: tables and → commands are exempt from the width limit.
func worstCase() *Report {
	const huge = 1023<<40 + 900<<30 // renders as "1023.9 TiB"
	r := goldenReport()
	r.Daemon.OperatingSystem = "Alpine Linux v3.24 (containerized)"
	r.Summary = Summary{DfReportedBytes: huge, DfImages: huge, DfContainers: huge, DfVolumes: huge, DfBuildCache: huge,
		LogBytes: huge, BuildCacheSharedBytes: huge, DockerBytesTotal: huge, NotShownByDf: huge, VolumeSizesUnknown: 999}
	r.BuildCache = BuildCache{Records: 99999, TotalBytes: huge, SharedBytes: huge, PrivateBytes: huge, ReclaimableBytes: huge,
		InUseBytes: huge, NotFreedByPrune: -1, SharedCountedInDfBytes: huge, DfRowMismatch: true, SharedMayBeFreedByPrune: true}
	r.Logs = Logs{TotalBytes: huge, Items: []LogItem{}}
	r.Projects = Projects{TotalBytes: huge, SharedImageBytes: huge, ImageOverlapBytes: huge, Items: []Project{}}
	r.Images = Images{Count: 99999, TotalBytes: huge, Dangling: 99999, ReclaimableBytes: huge, Items: []ImageItem{}}
	r.Volumes = Volumes{Count: 99999, TotalBytes: huge, UnknownSizes: 99999, Items: []VolumeItem{}}
	r.Filesystem = &Filesystem{Path: "/var/lib/docker", SizeBytes: huge, UsedBytes: huge, AvailableBytes: huge,
		DataRoot: &DiskUsage{AllocatedBytes: huge, Entries: 99999999, Complete: false, SkippedMounts: 999, OutsideDataRootBytes: huge, Items: []DiskItem{},
			Overlay2: &Overlay2Usage{LayerDB: true, ImageLayers: 99999, ImageLayerBytes: huge, ContainerLayers: 99999, ContainerLayerBytes: huge,
				Unreferenced: 99999, UnreferencedBytes: huge, NotExplainedByBuildCacheBytes: huge, LikelyLeftovers: true, LargestUnreferenced: []DiskItem{}}}}
	r.Desktop.ApparentBytes, r.Desktop.AllocatedBytes, r.Desktop.DockerReportedBytes = huge, huge, huge
	r.Desktop.Explanation = "Space inside the VM disk file is freed when images, containers, volumes or build cache are removed. The file itself only shrinks when Docker Desktop compacts it; lowering the disk usage limit under Settings → Resources → Advanced deletes the file and everything in it."
	r.Warnings = []string{"logs: 3 container(s) not measured: /var/lib/docker/containers is not visible from here: log directory not found under the data root. On Docker Desktop the data root is inside the VM; run the container image (see README)"}
	return r
}

func TestWriteTextProseFitsEightyColumns(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		r := worstCase()
		r.BuildCache.DfRowExcludesShared = legacy
		var buf bytes.Buffer
		if err := WriteText(&buf, r, 10); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(buf.String(), "1023.9 TiB") {
			t.Fatalf("worst case does not render 10-rune sizes:\n%s", buf.String())
		}
		for i, line := range strings.Split(buf.String(), "\n") {
			if strings.HasPrefix(line, "  → ") || strings.Contains(line, "com.docker.docker/Data/vms") {
				continue // copyable commands and a single path token are exempt
			}
			if f := strings.Fields(line); len(f) > 0 && units[f[0]] {
				t.Errorf("legacy=%v line %d starts with a unit split from its number:\n%s", legacy, i+1, line)
			}
			if n := runeLen(line); n > textWidth {
				t.Errorf("legacy=%v line %d has %d columns (max %d):\n%s", legacy, i+1, n, textWidth, line)
			}
		}
	}
}

func TestWrapTokensNeverSplitsATokenAndRespectsWidths(t *testing.T) {
	tokens := []string{"images 1023.9 TiB", "containers 1023.9 TiB", "volumes 1023.9 TiB", "build cache 1023.9 TiB"}
	got := wrapTokens(tokens, " · ", 40, 76)
	if strings.Join(got, " · ") != strings.Join(tokens, " · ") {
		t.Fatalf("wrapping lost or split a token: %q", got)
	}
	for i, l := range got {
		limit := 76
		if i == 0 {
			limit = 40
		}
		if runeLen(l) > limit {
			t.Errorf("line %d too long (%d > %d): %q", i, runeLen(l), limit, l)
		}
	}
	if long := wrapTokens([]string{strings.Repeat("x", 100)}, " ", 10, 10); len(long) != 1 || runeLen(long[0]) != 100 {
		t.Errorf("a token longer than the width must stay whole: %q", long)
	}
}

func TestWordsKeepsNumberAndUnitTogether(t *testing.T) {
	got := words("In use, never freed: 0 B and 140.7 MiB, (12.0 KiB) B-trees 3 B.")
	want := []string{"In", "use,", "never", "freed:", "0 B", "and", "140.7 MiB,", "(12.0 KiB)", "B-trees", "3 B."}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

// Values from the v0.3.1 demo on Engine 29.9 (docs/verification), where the
// wrapped build cache note ended a line with "0" and began the next with "B".
func TestWriteTextDemoValuesNeverSplitAUnit(t *testing.T) {
	r := goldenReport()
	r.BuildCache = BuildCache{Records: 4, TotalBytes: 147526000, SharedBytes: 147526000, PrivateBytes: 12288,
		ReclaimableBytes: 12288, NotFreedByPrune: -1, SharedMayBeFreedByPrune: true, SharedCountedInDfBytes: 147526000}
	var buf bytes.Buffer
	if err := WriteText(&buf, r, 10); err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(buf.String(), "\n") {
		if f := strings.Fields(line); len(f) > 0 && units[strings.TrimRight(f[0], ".,;:)")] {
			t.Errorf("line %d starts with a unit split from its number:\n%s", i+1, line)
		}
	}
}

func TestWriteTextBuildxBuilders(t *testing.T) {
	r := goldenReport()
	r.BuildCache.Builders = []BuildxBuilder{
		{Node: "ci0", Builder: "ci", Volume: "buildx_buildkit_ci0_state", Bytes: 7 << 30, Reference: "referenced"},
		{Node: "old0", Builder: "old", Volume: "buildx_buildkit_old0_state", Bytes: 300 << 20, Reference: "unreferenced"},
	}
	r.BuildCache.Suggestions = append(r.BuildCache.Suggestions, SuggestBuildxPrune("ci"))
	r.Volumes.Items = append(r.Volumes.Items, VolumeItem{Name: "buildx_buildkit_ci0_state", Driver: "local", Bytes: 7 << 30, RefCount: 1, Reference: "referenced", BuildxNode: "ci0"})
	var buf bytes.Buffer
	if err := WriteText(&buf, r, 0); err != nil {
		t.Fatal(err)
	}
	out := flat(buf.String())
	for _, want := range []string{
		"keep their own cache in a state volume. It is counted under VOLUMES, not in the figures above",
		"ci ci0 7.0 GiB builder container exists",
		"old old0 300.0 MiB no container: builder removed, state kept",
		"→ docker buildx prune --builder=ci",
		"referenced, buildx builder state (see BUILD CACHE)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, buf.String())
		}
	}
}

func TestSuggestBuildxPruneQuotesAndRefuses(t *testing.T) {
	if got := SuggestBuildxPrune("a b;rm -rf /"); !strings.HasPrefix(got, "docker buildx prune --builder='a b;rm -rf /'   #") {
		t.Fatalf("not quoted: %q", got)
	}
	if got := SuggestBuildxPrune("-x"); !strings.HasPrefix(got, "docker buildx prune --builder=-x ") && !strings.HasPrefix(got, "docker buildx prune --builder='-x'") {
		t.Fatalf("leading dash must stay attached to --builder=: %q", got)
	}
	for _, bad := range []string{"", "x\x1b[2J", "x‮"} {
		if got := SuggestBuildxPrune(bad); got != SuggestBuildxLs {
			t.Fatalf("%q: want the lookup hint, got %q", bad, got)
		}
	}
}

func TestWriteTextOverlay2Leftovers(t *testing.T) {
	r := goldenReport()
	r.Filesystem = &Filesystem{Path: "/var/lib/docker", SizeBytes: 100 << 30, UsedBytes: 90 << 30, AvailableBytes: 10 << 30,
		DataRoot: &DiskUsage{AllocatedBytes: 60 << 30, Complete: true, Items: []DiskItem{{Name: "overlay2", AllocatedBytes: 50 << 30}},
			Overlay2: &Overlay2Usage{LayerDB: true, ImageLayers: 40, ImageLayerBytes: 10 << 30, ContainerLayers: 2, ContainerLayerBytes: 1 << 20,
				Unreferenced: 3, UnreferencedBytes: 40 << 30, NotExplainedByBuildCacheBytes: 39 << 30, LikelyLeftovers: true,
				LargestUnreferenced: []DiskItem{{Name: "abcdefabcdefabcdefabcdefab", AllocatedBytes: 38 << 30}}}}}
	var buf bytes.Buffer
	if err := WriteText(&buf, r, 10); err != nil {
		t.Fatal(err)
	}
	out := flat(buf.String())
	for _, want := range []string{
		"image layers 10.0 GiB in 40 · container layers 1.0 MiB in 2", "no image or container record 40.0 GiB in 3",
		"39.0 GiB of these is more than all private build cache could explain",
		"abcdefabcdefabcdefabcdefab 38.0 GiB",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, buf.String())
		}
	}
	// Below the threshold: no finding, no table.
	r.Filesystem.DataRoot.Overlay2.LikelyLeftovers = false
	r.Filesystem.DataRoot.Overlay2.NotExplainedByBuildCacheBytes = 0
	buf.Reset()
	if err := WriteText(&buf, r, 10); err != nil {
		t.Fatal(err)
	}
	if out := flat(buf.String()); strings.Contains(out, "more than all private build cache") || strings.Contains(out, "LARGEST WITHOUT RECORD") ||
		!strings.Contains(out, "within what BuildKit snapshots") {
		t.Fatalf("no finding expected:\n%s", buf.String())
	}
}

func TestWriteTextOverlay2HiddenClassicStore(t *testing.T) {
	r := goldenReport()
	r.Filesystem = &Filesystem{Path: "/var/lib/docker", DataRoot: &DiskUsage{Complete: true,
		Items:    []DiskItem{{Name: "overlay2", AllocatedBytes: 12 << 30}},
		Overlay2: &Overlay2Usage{HiddenByContainerdStore: true, LargestUnreferenced: []DiskItem{}}}}
	var buf bytes.Buffer
	if err := WriteText(&buf, r, 10); err != nil {
		t.Fatal(err)
	}
	if out := flat(buf.String()); !strings.Contains(out, "overlay2: 12.0 GiB on disk from the classic image store") ||
		!strings.Contains(out, "Switch the daemon back to overlay2 to see and remove them") {
		t.Fatalf("hidden classic store not explained:\n%s", buf.String())
	}
}

func TestWriteTextClassicStoreHintWithoutDisk(t *testing.T) {
	r := goldenReport()
	r.Filesystem = &Filesystem{Path: "/var/lib/docker", SizeBytes: 100 << 30, UsedBytes: 10 << 30, AvailableBytes: 90 << 30, ClassicStoreLayers: 9}
	var buf bytes.Buffer
	if err := WriteText(&buf, r, 10); err != nil {
		t.Fatal(err)
	}
	if out := flat(buf.String()); !strings.Contains(out, "the classic image store still records 9 image layer(s) in overlay2/") {
		t.Fatalf("hint missing:\n%s", buf.String())
	}
}
