// Command dockdu explains where Docker's disk space went. It only reads:
// it never removes, stops or prunes anything.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

const usageText = `dockdu explains where Docker's disk space went and what prune would NOT give back.
It is read-only: nothing is removed, stopped or pruned.

Usage:
  dockdu [flags]

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
	usage := func(w io.Writer) {
		fmt.Fprint(w, usageText)
		fs.SetOutput(w)
		fs.PrintDefaults()
		fs.SetOutput(stderr)
	}
	fs.Usage = func() { usage(stderr) }
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout) // requested help goes to stdout so it can be piped
			return 0
		}
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
	fmt.Fprintln(stderr, "dockdu: the report is not implemented yet (phase 1 skeleton); run with --help")
	return 2
}
