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
	"runtime/debug"
	"strings"
	"time"

	"github.com/maximilianfeix/dockdu/internal/collect"
	"github.com/maximilianfeix/dockdu/internal/engine"
	"github.com/maximilianfeix/dockdu/internal/report"
)

// version is set at build time with -ldflags "-X main.version=...". Builds
// made with `go install module@vX.Y.Z` carry no ldflags; effectiveVersion
// falls back to the module version Go records in the binary.
var version = "dev"

func effectiveVersion() string {
	if version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			return strings.TrimPrefix(v, "v")
		}
	}
	return version
}

// exitUsage is EX_USAGE from sysexits; distinct from 2 (report with warnings).
const exitUsage = 64

const usageText = `dockdu explains where Docker's disk space went and what prune would NOT give back.
It is read-only: nothing is removed, stopped or pruned.

Usage:
  dockdu [flags]

Exit codes: 0 report complete · 1 cannot reach the Docker daemon · 2 report has warnings · 64 usage error

Flags:
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run parses args and returns the process exit code.
// Exit codes follow SPEC.md §5: 0 report complete, 1 daemon unreachable,
// 2 report rendered with warnings, 64 usage error (EX_USAGE), so a script
// can tell a typo from a degraded report.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("dockdu", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	asJSON := fs.Bool("json", false, "machine-readable output")
	top := fs.Int("top", 10, "rows per section in text output (0 = all)")
	dataRoot := fs.String("data-root", "", "override the Docker data root used to find log files (default: from the daemon)")
	noDesktop := fs.Bool("no-desktop", false, "skip the Docker Desktop / OrbStack section")
	disk := fs.Bool("disk", false, "also measure what the data root occupies on disk (walks every file with lstat; can take minutes on large hosts)")
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
		return exitUsage
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "dockdu: unexpected argument %q\n", fs.Arg(0))
		usage(stderr)
		return exitUsage
	}
	if *showVersion {
		fmt.Fprintf(stdout, "dockdu %s\n", effectiveVersion())
		return 0
	}
	// Development helper: record a sanitised snapshot as a test fixture.
	// Not a user-facing flag; it never overwrites an existing file.
	if path := os.Getenv("DOCKDU_RECORD_FIXTURE"); path != "" {
		return recordFixture(path, stderr)
	}

	eng, source, err := connect()
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
			fmt.Fprintf(stderr, "dockdu: cannot reach the Docker daemon at %s (from %s)\n  start Docker, or point DOCKER_HOST or DOCKER_CONTEXT at it; detail: %v\n", eng.Host(), source, err)
			return 1
		}
		fmt.Fprintf(stderr, "dockdu: %v\n", err)
		return 1
	}
	return render(snap, renderOpts{dataRoot: *dataRoot, desktop: !*noDesktop, disk: *disk, asJSON: *asJSON, top: *top}, stdout, stderr)
}

// renderOpts are the flags that shape a report.
type renderOpts struct {
	dataRoot      string
	desktop, disk bool
	asJSON        bool
	top           int
}

// diskWalkTimeout bounds --disk; a walk that runs out reports a lower bound.
const diskWalkTimeout = 10 * time.Minute

// render builds and writes the report; split from run so tests can feed a
// fixture snapshot through the same path.
func render(snap *engine.Snapshot, o renderOpts, stdout, stderr io.Writer) int {
	rep := collect.Build(snap, o.dataRoot, time.Now(), effectiveVersion())
	root := rep.Daemon.DataRoot
	if root == "" {
		root = "/var/lib/docker"
	}
	ctx, cancel := context.WithTimeout(context.Background(), diskWalkTimeout)
	fsys, w := collect.Filesystem(ctx, root, o.disk, collect.DefaultDiskBudget)
	cancel()
	rep.Filesystem = fsys
	collect.InterpretOverlay2(fsys, rep)
	if fsys != nil && rep.Daemon.ImageStore == "containerd" {
		fsys.ClassicStoreLayers = collect.ClassicStoreLayers(root)
		if fsys.ClassicStoreLayers > 0 && fsys.DataRoot == nil {
			rep.Summary.Suggestions = append(rep.Summary.Suggestions, report.SuggestDiskWalk)
		}
	}
	rep.Warnings = append(rep.Warnings, w...)
	if fsys != nil && fsys.DataRoot == nil && fsys.ClassicStoreLayers == 0 && fsys.SizeBytes > 0 && fsys.UsedBytes*5 >= fsys.SizeBytes*4 {
		// The filesystem is at least 80 % full: say how to find out whether
		// the space is inside the data root.
		rep.Summary.Suggestions = append(rep.Summary.Suggestions, report.SuggestDiskWalk)
	}
	if o.desktop && runtime.GOOS == "darwin" {
		home, err := os.UserHomeDir()
		if err != nil {
			rep.Warnings = append(rep.Warnings, "desktop: home directory unknown: "+err.Error())
		} else {
			d, w := collect.Desktop(home, rep.Summary.DockerBytesTotal)
			rep.Desktop = d
			rep.Warnings = append(rep.Warnings, w...)
		}
	}
	var err error
	if o.asJSON {
		err = report.WriteJSON(stdout, rep)
	} else {
		err = report.WriteText(stdout, rep, o.top)
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

// connect resolves the daemon the docker CLI would use (DOCKER_HOST, then
// DOCKER_CONTEXT, then the CLI's current context) and returns a client
// for it plus a description of where the address came from.
func connect() (*engine.Moby, string, error) {
	home, _ := os.UserHomeDir() // empty home only disables context lookup
	target, err := engine.ResolveTarget(os.Getenv, home)
	if err != nil {
		return nil, "", err
	}
	eng, err := engine.NewMoby(target.Opts...)
	if err != nil {
		return nil, "", err
	}
	return eng, target.Source, nil
}

func recordFixture(path string, stderr io.Writer) int {
	eng, _, err := connect()
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
