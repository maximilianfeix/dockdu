package collect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilianfeix/dockdu/internal/engine"
	"github.com/maximilianfeix/dockdu/internal/report"
)

func writeFile(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLogsJSONFileSumsRotatedSiblingsAndFlagsRotation(t *testing.T) {
	root := t.TempDir()
	id := "1f1412f2177f1838b0d6e8b7b91dfab5547324fdfc0a00a105e9acbf8c44fd87"
	dir := filepath.Join(root, "containers", id)
	writeFile(t, filepath.Join(dir, id+"-json.log"), 1000)
	writeFile(t, filepath.Join(dir, id+"-json.log.1"), 1024)
	writeFile(t, filepath.Join(dir, id+"-json.log.2.gz"), 100)
	writeFile(t, filepath.Join(dir, "config.v2.json"), 50)     // not a log
	writeFile(t, filepath.Join(dir, id+"-json.log.bak"), 5000) // not a rotation name
	s := &engine.Snapshot{Daemon: engine.Daemon{DataRoot: root}, Containers: []engine.Container{
		{ID: id, Name: "shop-web-1", LogDriver: "json-file", LogPath: filepath.Join(dir, id+"-json.log"),
			LogOpts: map[string]string{"max-size": "1k", "max-file": "3"},
			Labels:  map[string]string{"com.docker.compose.project": "shop"}},
	}}
	got, warns := Logs(s, root)
	if len(warns) != 0 {
		t.Fatalf("warnings: %v", warns)
	}
	if !got.Complete || got.TotalBytes != 2124 {
		t.Fatalf("complete=%v total=%d, want true/2124", got.Complete, got.TotalBytes)
	}
	it := got.Items[0]
	if it.Files != 3 || it.Bytes != 2124 || !it.Measured || it.Rotation != "bounded" || it.MaxSize != "1k" || it.MaxFile != "3" || it.Project != "shop" {
		t.Fatalf("item = %+v", it)
	}
	if len(got.Suggestions) != 0 {
		t.Fatalf("bounded logs need no suggestion, got %v", got.Suggestions)
	}
}

func TestLogsJSONFileUnboundedDerivesPathAndSuggests(t *testing.T) {
	root := t.TempDir()
	id := "871547ac7ec5f8c01acd304332b3716566039690ea073703204a6d8f3c1ca29d"
	writeFile(t, filepath.Join(root, "containers", id, id+"-json.log"), 85)
	// LogPath left empty on purpose: dockdu must derive it from the data root.
	s := &engine.Snapshot{Containers: []engine.Container{{ID: id, Name: "plain", LogDriver: "json-file"}}}
	got, _ := Logs(s, root)
	it := got.Items[0]
	if it.Rotation != "unbounded" || it.Bytes != 85 || !it.Measured {
		t.Fatalf("item = %+v", it)
	}
	if len(got.Suggestions) == 0 || !strings.Contains(got.Suggestions[0], "max-size") {
		t.Fatalf("unbounded json-file log should suggest log-opts, got %v", got.Suggestions)
	}
}

func TestLogsLocalDriverUsesLocalLogsDir(t *testing.T) {
	root := t.TempDir()
	id := "34b1d5c11f0415a5f9a28c2b9b08128f8e5671f07c5041b088e8fd39514e1c2e"
	dir := filepath.Join(root, "containers", id, "local-logs")
	writeFile(t, filepath.Join(dir, "container.log"), 2740)
	writeFile(t, filepath.Join(dir, "container.log.1.gz"), 300)
	s := &engine.Snapshot{Containers: []engine.Container{{ID: id, Name: "loc", LogDriver: "local"}}}
	got, _ := Logs(s, root)
	it := got.Items[0]
	if it.Bytes != 3040 || it.Files != 2 || it.Rotation != "bounded" || !it.Measured {
		t.Fatalf("item = %+v", it)
	}
	if it.MaxSize != "20m (driver default)" || it.MaxFile != "5 (driver default)" {
		t.Fatalf("local defaults must be labelled as defaults, got %q/%q", it.MaxSize, it.MaxFile)
	}
}

func TestLogsSymlinksAreNotFollowed(t *testing.T) {
	root := t.TempDir()
	id := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	dir := filepath.Join(root, "containers", id)
	writeFile(t, filepath.Join(dir, id+"-json.log"), 10)
	big := filepath.Join(t.TempDir(), "big")
	writeFile(t, big, 100000)
	if err := os.Symlink(big, filepath.Join(dir, id+"-json.log.1")); err != nil {
		t.Skip("symlinks not supported here")
	}
	s := &engine.Snapshot{Containers: []engine.Container{{ID: id, Name: "sym", LogDriver: "json-file"}}}
	got, _ := Logs(s, root)
	if got.Items[0].Bytes != 10 || got.Items[0].Files != 1 {
		t.Fatalf("symlinked file must be ignored, got %+v", got.Items[0])
	}
}

func TestLogsMissingDirIsUnmeasuredNotZero(t *testing.T) {
	root := t.TempDir()
	s := &engine.Snapshot{Containers: []engine.Container{{ID: "bbbbbbbbbbbb", Name: "gone", LogDriver: "json-file"}}}
	got, warns := Logs(s, root)
	it := got.Items[0]
	if it.Measured || it.Note == "" || got.Complete || got.TotalBytes != 0 || it.Bytes != -1 {
		t.Fatalf("unmeasured container must be flagged with bytes -1, got item=%+v complete=%v", it, got.Complete)
	}
	if len(warns) != 1 {
		t.Fatalf("one warning expected, got %v", warns)
	}
}

func TestLogsOtherDriverIsNotOnDisk(t *testing.T) {
	root := t.TempDir()
	s := &engine.Snapshot{Containers: []engine.Container{{ID: "cccccccccccc", Name: "jd", LogDriver: "journald"}}}
	got, warns := Logs(s, root)
	it := got.Items[0]
	if it.Rotation != "not on disk" || it.Measured || !got.Complete || len(warns) != 0 {
		t.Fatalf("other drivers are not measured and do not make the report incomplete: %+v %v", it, warns)
	}
}

func TestLogsInspectErrorIsUnmeasured(t *testing.T) {
	s := &engine.Snapshot{Containers: []engine.Container{{ID: "dddddddddddd", Name: "x", InspectError: "boom"}}}
	got, _ := Logs(s, t.TempDir())
	if got.Items[0].Measured || got.Complete || !strings.Contains(got.Items[0].Note, "inspect") {
		t.Fatalf("got %+v", got.Items[0])
	}
}

func TestLogsSortedBySizeDescThenName(t *testing.T) {
	root := t.TempDir()
	ids := map[string]int{"aaaaaaaaaaaa": 10, "bbbbbbbbbbbb": 30, "cccccccccccc": 30}
	var cs []engine.Container
	for id, n := range ids {
		writeFile(t, filepath.Join(root, "containers", id, id+"-json.log"), n)
		cs = append(cs, engine.Container{ID: id, Name: id[:1], LogDriver: "json-file"})
	}
	got, _ := Logs(&engine.Snapshot{Containers: cs}, root)
	if got.Items[0].Name != "b" || got.Items[1].Name != "c" || got.Items[2].Name != "a" {
		t.Fatalf("order wrong: %v %v %v", got.Items[0].Name, got.Items[1].Name, got.Items[2].Name)
	}
}

func TestLogsEmpty(t *testing.T) {
	got, warns := Logs(&engine.Snapshot{}, t.TempDir())
	if !got.Complete || got.Items == nil || len(got.Items) != 0 || got.Suggestions == nil || len(warns) != 0 {
		t.Fatalf("empty: %+v %v", got, warns)
	}
}

func TestLogsDataRootOverrideRebasesReportedPath(t *testing.T) {
	root := t.TempDir()
	id := "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	writeFile(t, filepath.Join(root, "containers", id, id+"-json.log"), 42)
	s := &engine.Snapshot{Daemon: engine.Daemon{DataRoot: "/var/lib/docker"}, Containers: []engine.Container{
		{ID: id, Name: "moved", LogDriver: "json-file", LogPath: "/var/lib/docker/containers/" + id + "/" + id + "-json.log"},
	}}
	got, _ := Logs(s, root)
	if !got.Items[0].Measured || got.Items[0].Bytes != 42 {
		t.Fatalf("path under the daemon root must be rebased onto --data-root, got %+v", got.Items[0])
	}
	// Without an override the daemon's absolute path is used as is.
	got, _ = Logs(s, "")
	if got.Items[0].Measured {
		t.Fatalf("without override the daemon path (missing here) must stay unmeasured, got %+v", got.Items[0])
	}
}

func TestLogsRejectsInvalidIDsAndPathsOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	writeFile(t, filepath.Join(outside, "x-json.log"), 999)
	s := &engine.Snapshot{Daemon: engine.Daemon{DataRoot: root}, Containers: []engine.Container{
		{ID: "../../etc", Name: "evil-id", LogDriver: "json-file"},
		{ID: "ffffffffffff", Name: "evil-path", LogDriver: "json-file", LogPath: filepath.Join(outside, "x-json.log")},
	}}
	got, warns := Logs(s, "")
	for _, it := range got.Items {
		if it.Measured {
			t.Fatalf("must not measure %s: %+v", it.Name, it)
		}
	}
	if got.Complete || len(warns) != 2 {
		t.Fatalf("both must be flagged: complete=%v warns=%v", got.Complete, warns)
	}
}

func TestLogsInvisibleDataRootWarnsOnce(t *testing.T) {
	root := filepath.Join(t.TempDir(), "inside-the-vm") // does not exist on this host
	s := &engine.Snapshot{Containers: []engine.Container{
		{ID: "aaaaaaaaaaaa", Name: "a", LogDriver: "json-file"},
		{ID: "bbbbbbbbbbbb", Name: "b", LogDriver: "local"},
		{ID: "cccccccccccc", Name: "c", LogDriver: "journald"},
	}}
	got, warns := Logs(s, root)
	if len(warns) != 1 || !strings.Contains(warns[0], "2 container(s) not measured") {
		t.Fatalf("want one summarising warning, got %v", warns)
	}
	if got.Complete || got.Items[0].Measured || got.Items[0].Bytes != -1 || got.Items[0].Note == "" {
		t.Fatalf("items must stay unmeasured with a note: %+v", got.Items)
	}
}

// A container directory replaced by a symlink to a directory outside the
// data root must not be listed: dockdu would otherwise report sizes of
// files it was never meant to see (security audit S3).
func TestLogsDoNotFollowContainerDirSymlinkOutOfRoot(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "docker")
	outside := filepath.Join(base, "elsewhere")
	id := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	writeFile(t, filepath.Join(outside, id+"-json.log"), 4096)
	if err := os.MkdirAll(filepath.Join(root, "containers"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "containers", id)); err != nil {
		t.Fatal(err)
	}
	s := &engine.Snapshot{Containers: []engine.Container{{ID: id, Name: "swapped", LogDriver: "json-file"}}}
	got, _ := Logs(s, root)
	if it := got.Items[0]; it.Measured || it.Bytes != -1 || got.Complete {
		t.Fatalf("a symlink escaping the data root must leave the container unmeasured, got %+v", it)
	}
}

// Observed on docker:29-dind-rootless: a rootless daemon reads
// ~/.config/docker/daemon.json, so the log-opts hint must name that file.
func TestLogsRootlessDaemonSuggestsUserDaemonJSON(t *testing.T) {
	root := t.TempDir()
	id := "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	writeFile(t, filepath.Join(root, "containers", id, id+"-json.log"), 100)
	for _, rootless := range []bool{false, true} {
		s := &engine.Snapshot{Daemon: engine.Daemon{Rootless: rootless},
			Containers: []engine.Container{{ID: id, Name: "chatty", LogDriver: "json-file"}}}
		got, _ := Logs(s, root)
		want := report.SuggestLogOptsDaemon
		if rootless {
			want = report.SuggestLogOptsDaemonRootless
		}
		if len(got.Suggestions) == 0 || got.Suggestions[0] != want {
			t.Fatalf("rootless=%v: suggestions %q", rootless, got.Suggestions)
		}
	}
}
