package main

import (
	"bytes"
	"strings"
	"testing"
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

func TestRunNoArgsNotImplementedYet(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run(nil, &out, &errOut); code != 2 {
		t.Fatalf("exit code = %d, want 2 until the report exists", code)
	}
}
