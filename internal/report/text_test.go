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
			LogBytes: 51861225472, LogBytesComplete: false, BuildCacheSharedBytes: 33285996544, DockerBytesTotal: 108979175424, NotShownByDf: 51861225472, Suggestions: []string{}},
		BuildCache: BuildCache{Records: 26, TotalBytes: 68824252416, SharedBytes: 33285996544, PrivateBytes: 35538255872, ReclaimableBytes: 35538255872, NotFreedByPrune: 33285996544,
			Suggestions: []string{SuggestBuilderPruneOld}},
		Logs: Logs{TotalBytes: 51861225472, Complete: false, Items: []LogItem{
			{ContainerID: "a", Name: "shop-web-1", Project: "shop", Driver: "json-file", Rotation: "unbounded", Bytes: 6657199104, Files: 1, Measured: true},
			{ContainerID: "b", Name: "blog-worker-3", Project: "blog", Driver: "local", Rotation: "bounded", MaxSize: "20m (driver default)", MaxFile: "5 (driver default)", Bytes: 104857600, Files: 5, Measured: true},
			{ContainerID: "c", Name: "loose", Driver: "json-file", Rotation: "unbounded", Bytes: -1, Measured: false, Note: "log directory not readable (run as root or inside the container image)"},
		}, Suggestions: []string{SuggestLogOptsDaemon}},
		// TotalBytes = Σ items (23837068287 + 15073741824) + shared 500000000.
		Projects: Projects{TotalBytes: 39410810111, SharedImageBytes: 500000000, Items: []Project{
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
	if !strings.Contains(out, "IMAGES  (top 1 of 3") {
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
	var buf bytes.Buffer
	if err := WriteText(&buf, r, 10); err != nil {
		t.Fatal(err)
	}
	if out := buf.String(); !strings.Contains(out, "= df + logs (df's build cache row already leaves out 31.0 GiB shared with images)") || strings.Contains(out, "= df − 0 B") {
		t.Fatalf("legacy summary line wrong:\n%s", out)
	}
}

func TestSuggestVolumeRmQuotesDaemonNames(t *testing.T) {
	cases := map[string]string{
		"pgdata":          "docker volume rm pgdata   #",
		"x; rm -rf ~":     "docker volume rm 'x; rm -rf ~'   #",
		"it's":            `docker volume rm 'it'\''s'   #`,
		"$(touch /tmp/p)": "docker volume rm '$(touch /tmp/p)'   #",
		"-rf":             "docker volume rm '-rf'   #",
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
