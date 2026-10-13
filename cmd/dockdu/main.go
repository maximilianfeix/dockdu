// Command dockdu explains where Docker's disk space went. It only reads:
// it never removes, stops or prunes anything.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/maximilianfeix/dockdu/internal/collect"
	"github.com/maximilianfeix/dockdu/internal/engine"
	"github.com/maximilianfeix/dockdu/internal/report"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

const usageText = `dockdu explains where Docker's disk space went and what prune would NOT give back.
It is read-only: nothing is removed, stopped or pruned.

Usage:
  dockdu [flags]

Exit codes: 0 report complete · 1 cannot reach the Docker daemon · 2 report has warnings

Flags:
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run parses args and returns the process exit code.
// Exit codes follow SPEC.md §5: 0 report complete, 1 daemon unreachable,
// 2 report rendered with warnings. Usage errors also exit 2 (Go flag
// convention; not defined by the spec).
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("dockdu", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	asJSON := fs.Bool("json", false, "machine-readable output")
	top := fs.Int("top", 10, "rows per section in text output (0 = all)")
	dataRoot := fs.String("data-root", "", "override the Docker data root used to find log files (default: from the daemon)")
	noDesktop := fs.Bool("no-desktop", false, "skip the Docker Desktop / OrbStack section")
	usage := func(w io.Writer) {
		fmt.Fprint(w, usageText)
		fs.SetOutput(w)
		fs.PrintDefaults()
		fs.SetOutput(stderr)
	}
	// flag calls Usage itself on -h and on errors; keep that silent and
	// print usage explicitly so help goes to stdout and errors to stderr.
	fs.Usage = func() {}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout) // requested help goes to stdout so it can be piped
			return 0
		}
		usage(stderr) // flag already printed the error line
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "dockdu: unexpected argument %q\n", fs.Arg(0))
		usage(stderr)
		return 2
	}
	if *showVersion {
		fmt.Fprintf(stdout, "dockdu %s\n", version)
		return 0
	}
	// Development helper: record a sanitised snapshot as a test fixture.
	// Not a user-facing flag; it never overwrites an existing file.
	if path := os.Getenv("DOCKDU_RECORD_FIXTURE"); path != "" {
		return recordFixture(path, stderr)
	}

	eng, err := engine.NewMoby()
	if err != nil {
		fmt.Fprintf(stderr, "dockdu: %v\n", err)
		return 1
	}
	defer eng.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	snap, err := eng.Snapshot(ctx)
	if err != nil {
		if errors.Is(err, engine.ErrDaemonUnreachable) {
			fmt.Fprintf(stderr, "dockdu: cannot reach the Docker daemon at %s\n  set DOCKER_HOST or start Docker; detail: %v\n", eng.Host(), err)
			return 1
		}
		fmt.Fprintf(stderr, "dockdu: %v\n", err)
		return 1
	}
	return render(snap, *dataRoot, !*noDesktop, *asJSON, *top, stdout, stderr)
}

// render builds and writes the report; split from run so tests can feed a
// fixture snapshot through the same path.
func render(snap *engine.Snapshot, dataRoot string, desktop, asJSON bool, top int, stdout, stderr io.Writer) int {
	rep := collect.Build(snap, dataRoot, time.Now(), version)
	if desktop && runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		d, w := collect.Desktop(home, rep.Summary.DockerBytesTotal)
		rep.Desktop = d
		rep.Warnings = append(rep.Warnings, w...)
	}
	var err error
	if asJSON {
		err = report.WriteJSON(stdout, rep)
	} else {
		err = report.WriteText(stdout, rep, top)
	}
	if err != nil {
		fmt.Fprintf(stderr, "dockdu: write report: %v\n", err)
		return 2
	}
	if len(rep.Warnings) > 0 {
		return 2
	}
	return 0
}

func recordFixture(path string, stderr io.Writer) int {
	eng, err := engine.NewMoby()
	if err != nil {
		fmt.Fprintf(stderr, "dockdu: %v\n", err)
		return 1
	}
	defer eng.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := engine.Record(ctx, eng, path, false); err != nil {
		if errors.Is(err, engine.ErrDaemonUnreachable) {
			fmt.Fprintf(stderr, "dockdu: cannot reach the Docker daemon at %s: %v\n", eng.Host(), err)
			return 1
		}
		fmt.Fprintf(stderr, "dockdu: record fixture: %v\n", err)
		return 2
	}
	fmt.Fprintf(stderr, "dockdu: fixture written to %s\n", path)
	return 0
}
