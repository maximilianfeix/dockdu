package main

import (
	"bytes"
	"encoding/json"
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
	if code := run([]string{"--version", "extra"}, &out, &errOut); code != 2 {
		t.Fatalf("exit code = %d, want 2 for a positional argument", code)
	}
	if !strings.Contains(errOut.String(), "unexpected argument") {
		t.Fatalf("stderr should name the argument, got %q", errOut.String())
	}
}

func TestRunUnknownFlag(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"--bogus"}, &out, &errOut); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
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
