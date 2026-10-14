package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/maximilianfeix/dockdu/internal/engine"
)

func TestRunVersion(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"--version"}, &out, &errOut); code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, errOut.String())
	}
	if got := out.String(); got != "dockdu dev\n" {
		t.Fatalf("stdout = %q, want %q", got, "dockdu dev\n")
	}
}

func TestRunHelp(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"--help"}, &out, &errOut); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "read-only") {
		t.Fatalf("help must go to stdout and state that dockdu is read-only, got stdout=%q stderr=%q", out.String(), errOut.String())
	}
	if errOut.Len() != 0 {
		t.Fatalf("help must not write to stderr, got %q", errOut.String())
	}
}

func TestRunRejectsPositionalArguments(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"--version", "extra"}, &out, &errOut); code != exitUsage {
		t.Fatalf("exit code = %d, want %d for a positional argument", code, exitUsage)
	}
	if !strings.Contains(errOut.String(), "unexpected argument") {
		t.Fatalf("stderr should name the argument, got %q", errOut.String())
	}
}

func TestRunUnknownFlag(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"--bogus"}, &out, &errOut); code != exitUsage {
		t.Fatalf("exit code = %d, want %d", code, exitUsage)
	}
}

func TestRunUnreachableDaemonExits1(t *testing.T) {
	t.Setenv("DOCKER_HOST", "unix:///nonexistent/dockdu-test.sock")
	var out, errOut bytes.Buffer
	if code := run(nil, &out, &errOut); code != 1 {
		t.Fatalf("exit code = %d, want 1; stderr=%q", code, errOut.String())
	}
	if !strings.Contains(errOut.String(), "DOCKER_HOST") || !strings.Contains(errOut.String(), "nonexistent/dockdu-test.sock") {
		t.Fatalf("stderr should name DOCKER_HOST and the socket, got %q", errOut.String())
	}
	if out.Len() != 0 {
		t.Fatalf("no report on stdout when the daemon is unreachable, got %q", out.String())
	}
}

func TestRenderFromFixtureExitCodes(t *testing.T) {
	s, err := engine.LoadFixture("../../internal/collect/testdata/desktop-29.2.json")
	if err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	// Log directories are not available here, so the report carries warnings → 2.
	if code := render(s, t.TempDir(), false, false, 10, &out, &errOut); code != 2 {
		t.Fatalf("exit code = %d, want 2 with warnings; stderr=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "SUMMARY") || !strings.Contains(out.String(), "WARNINGS") {
		t.Fatalf("text report incomplete: %q", out.String())
	}
	out.Reset()
	if code := render(s, t.TempDir(), false, true, 10, &out, &errOut); code != 2 {
		t.Fatalf("json exit code = %d, want 2", code)
	}
	var doc map[string]any
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("json output invalid: %v", err)
	}
	if _, ok := doc["summary"].(map[string]any)["docker_bytes_total"]; !ok {
		t.Fatalf("summary.docker_bytes_total missing in %v", doc["summary"])
	}
}

func TestRenderExitZeroWhenEverythingIsMeasured(t *testing.T) {
	s, err := engine.LoadFixture("../../internal/collect/testdata/desktop-29.2.json")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, c := range s.Containers {
		dir := filepath.Join(root, "containers", c.ID)
		if c.LogDriver == "local" {
			dir = filepath.Join(dir, "local-logs")
			mustWrite(t, filepath.Join(dir, "container.log"), 10)
			continue
		}
		mustWrite(t, filepath.Join(dir, c.ID+"-json.log"), 10)
	}
	var out, errOut bytes.Buffer
	if code := render(s, root, false, true, 10, &out, &errOut); code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q stdout=%s", code, errOut.String(), out.String())
	}
	var doc struct {
		Daemon struct {
			DataRoot         string `json:"data_root"`
			ReportedDataRoot string `json:"reported_data_root"`
		} `json:"daemon"`
		Warnings []string `json:"warnings"`
		Logs     struct {
			Complete bool `json:"complete"`
		} `json:"logs"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Warnings) != 0 || !doc.Logs.Complete {
		t.Fatalf("expected a complete report, got warnings=%v complete=%v", doc.Warnings, doc.Logs.Complete)
	}
	if doc.Daemon.DataRoot != root || doc.Daemon.ReportedDataRoot != "/var/lib/docker" {
		t.Fatalf("effective data root must be reported: %+v", doc.Daemon)
	}
}

func mustWrite(t *testing.T, path string, n int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, n), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEffectiveVersionFallsBackToBuildInfo(t *testing.T) {
	// In `go test` the main module version is "(devel)" or empty, so the
	// ldflags default must survive unchanged.
	if got := effectiveVersion(); got != "dev" {
		t.Fatalf("effectiveVersion() = %q, want %q under go test", got, "dev")
	}
	version = "1.2.3"
	t.Cleanup(func() { version = "dev" })
	if got := effectiveVersion(); got != "1.2.3" {
		t.Fatalf("ldflags version must win, got %q", got)
	}
}

// jsonPaths collects every key path of a decoded JSON document, with list
// elements written as [], and fails on any null value.
func jsonPaths(t *testing.T, prefix string, v any, out map[string]bool) {
	t.Helper()
	switch x := v.(type) {
	case nil:
		t.Errorf("%s is null; lists must be [] and objects omitted", prefix)
	case map[string]any:
		for k, e := range x {
			p := prefix + "." + k
			out[p] = true
			if strings.HasSuffix(k, "bytes") || strings.HasSuffix(k, "_bytes_total") {
				if f, ok := e.(float64); !ok || f != float64(int64(f)) || f < -1 {
					t.Errorf("%s = %v: byte values are integers >= -1", p, e)
				}
			}
			jsonPaths(t, p, e, out)
		}
	case []any:
		for _, e := range x {
			jsonPaths(t, prefix+"[]", e, out)
		}
	}
}

// TestJSONContract guards the --json data contract: stdout is exactly one
// JSON object, nothing goes to stderr, no value is null, byte values are
// integers, and no key that existed in a previous release disappears.
// testdata/json-keys.golden lists every key path; new keys may be added
// (run with -update-keys), removing one needs a schema_version bump.
func TestJSONContract(t *testing.T) {
	s, err := engine.LoadFixture("../../internal/collect/testdata/desktop-29.2.json")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, c := range s.Containers {
		mustWrite(t, filepath.Join(root, "containers", c.ID, c.ID+"-json.log"), 10)
	}
	var out, errOut bytes.Buffer
	render(s, root, false, true, 10, &out, &errOut)
	if errOut.Len() != 0 {
		t.Fatalf("--json must keep stderr empty for a rendered report, got %q", errOut.String())
	}
	dec := json.NewDecoder(&out)
	var doc map[string]any
	if err := dec.Decode(&doc); err != nil {
		t.Fatal(err)
	}
	if dec.More() {
		t.Fatal("stdout must hold exactly one JSON document")
	}
	if doc["schema_version"] != float64(1) {
		t.Fatalf("schema_version = %v, want 1", doc["schema_version"])
	}
	paths := map[string]bool{}
	jsonPaths(t, "", doc, paths)

	const golden = "testdata/json-keys.golden"
	if os.Getenv("DOCKDU_UPDATE_KEYS") == "1" {
		var keys []string
		for k := range paths {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		if err := os.WriteFile(golden, []byte(strings.Join(keys, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v (create it with DOCKDU_UPDATE_KEYS=1)", err)
	}
	for _, k := range strings.Fields(string(want)) {
		if !paths[k] {
			t.Errorf("key %s disappeared from the JSON output; that breaks scripts (bump SchemaVersion if intended)", k)
		}
	}
}
