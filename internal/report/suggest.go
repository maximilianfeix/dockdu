package report

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Suggestions are printed as text and never executed. None of them touches
// volumes: a volume holds data that no other Docker object reproduces.
const (
	SuggestBuilderPruneOld    = "docker builder prune --filter until=168h   # unused build cache older than 7 days"
	SuggestBuilderPruneAll    = "docker builder prune -a                    # all build cache not in use (next build is slower)"
	SuggestImagePruneDangling = "docker image prune                         # untagged (<none>) images only"
	SuggestImagePruneAll      = "docker image prune -a                      # removes EVERY image no container uses, including locally built ones that cannot be re-pulled"
	SuggestLogOptsDaemon      = `set "log-opts" in /etc/docker/daemon.json: {"log-driver":"json-file","log-opts":{"max-size":"10m","max-file":"3"}}   # new containers only`
	SuggestDiskWalk           = "dockdu --disk   # read-only: measures what the data root occupies on disk, to tell Docker's share of this filesystem from everything else (walks every file; can take minutes)"
	SuggestLogOptsContainer   = "recreate the container with --log-opt max-size=10m --log-opt max-file=3   # or --log-driver local"
)

// SuggestVolumeRm formats the single-volume removal hint. There is no
// bulk form on purpose. The name comes from the daemon, so it is quoted
// for a POSIX shell; a name with control characters gets no command at
// all, because no quoting makes such a line safe to paste.
func SuggestVolumeRm(name string) string {
	if strings.IndexFunc(name, unsafeRune) >= 0 {
		return "# volume " + strconv.QuoteToASCII(name) + " is unreferenced; its name contains control or invisible formatting characters, check `docker volume ls` before removing anything"
	}
	arg := ShellQuote(name)
	if strings.HasPrefix(name, "-") {
		arg = "-- " + arg // quoting does not stop a leading dash being read as a flag
	}
	return "docker volume rm " + arg + "   # only if you are sure nothing needs its data"
}

// unsafeRune reports characters that can change how a terminal shows a
// line: control characters (Cc), invisible format characters such as bidi
// overrides (Cf), and line or paragraph separators (Zl, Zp).
func unsafeRune(r rune) bool {
	return unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zl, unicode.Zp)
}

// shellSafe is the set of names that need no quoting. Docker's local
// volume driver only accepts names of this form; plugins may not.
var shellSafe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

// ShellQuote returns s unchanged when it is a plain name and otherwise
// wraps it in single quotes, escaping embedded single quotes, so a crafted
// name cannot add a second command to a suggestion.
func ShellQuote(s string) string {
	if shellSafe.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
