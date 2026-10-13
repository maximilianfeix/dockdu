package collect

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/maximilianfeix/dockdu/internal/engine"
	"github.com/maximilianfeix/dockdu/internal/report"
)

// composeProjectLabel is the compose label that names a project.
const composeProjectLabel = "com.docker.compose.project"

// rotatedSuffix matches "", ".1", ".12", ".1.gz" after the base log name.
var rotatedSuffix = regexp.MustCompile(`^(\.[0-9]+(\.gz)?)?$`)

// containerID is the daemon's 64-hex (or shortened) identifier. Anything
// else is not joined into a path.
var containerID = regexp.MustCompile(`^[0-9a-f]{12,64}$`)

// Logs applies SPEC §6.5. It only stats files: it never opens, reads or
// follows symbolic links. dataRoot overrides the daemon's data root (the
// snapshot value is used when dataRoot is empty); with an override, log
// paths the daemon reported under its own root are rebased onto it, which
// is what makes `--data-root /mnt/docker` work when the root is mounted
// somewhere else.
func Logs(s *engine.Snapshot, dataRoot string) (report.Logs, []string) {
	daemonRoot := s.Daemon.DataRoot
	if daemonRoot == "" {
		daemonRoot = "/var/lib/docker"
	}
	override := dataRoot != "" && filepath.Clean(dataRoot) != filepath.Clean(daemonRoot)
	if dataRoot == "" {
		dataRoot = daemonRoot
	}
	out := report.Logs{Complete: true, Items: make([]report.LogItem, 0, len(s.Containers)), Suggestions: []string{}}
	var warns []string
	unbounded := false

	for _, c := range s.Containers {
		it := report.LogItem{ContainerID: c.ID, Name: c.Name, Project: c.Labels[composeProjectLabel], Driver: c.LogDriver}
		switch {
		case c.InspectError != "":
			it.Rotation = "unknown"
			it.Note = "inspect failed: " + c.InspectError
			out.Complete = false
			warns = append(warns, fmt.Sprintf("logs: %s not measured, inspect failed", c.Name))
		case !containerID.MatchString(c.ID):
			it.Rotation = "unknown"
			it.Note = "invalid container id from the daemon; not used in a path"
			out.Complete = false
			warns = append(warns, fmt.Sprintf("logs: %s not measured: invalid container id %q", c.Name, c.ID))
		case c.LogDriver == "json-file":
			it.Rotation, it.MaxSize, it.MaxFile = rotationJSONFile(c.LogOpts)
			path := c.LogPath
			switch {
			case path == "":
				path = filepath.Join(dataRoot, "containers", c.ID, c.ID+"-json.log")
			case override && strings.HasPrefix(path, daemonRoot+string(filepath.Separator)):
				path = filepath.Join(dataRoot, strings.TrimPrefix(path, daemonRoot))
			}
			if !underRoot(path, dataRoot) {
				it.Note = "log path reported by the daemon is outside the data root; not measured"
				out.Complete = false
				warns = append(warns, fmt.Sprintf("logs: %s not measured: path outside data root", c.Name))
				break
			}
			measure(&it, filepath.Dir(path), filepath.Base(path), &out, &warns)
			if it.Measured && it.Rotation == "unbounded" && it.Bytes > 0 {
				unbounded = true
			}
		case c.LogDriver == "local":
			it.Rotation = "bounded"
			it.MaxSize = optOrDefault(c.LogOpts, "max-size", "20m (driver default)")
			it.MaxFile = optOrDefault(c.LogOpts, "max-file", "5 (driver default)")
			measure(&it, filepath.Join(dataRoot, "containers", c.ID, "local-logs"), "container.log", &out, &warns)
		default:
			it.Rotation = "not on disk"
			it.Note = fmt.Sprintf("driver %q does not write under the data root", c.LogDriver)
		}
		out.Items = append(out.Items, it)
	}

	sort.SliceStable(out.Items, func(i, j int) bool {
		a, b := out.Items[i], out.Items[j]
		if a.Measured != b.Measured {
			return a.Measured
		}
		if a.Bytes != b.Bytes {
			return a.Bytes > b.Bytes
		}
		return a.Name < b.Name
	})
	if unbounded {
		out.Suggestions = append(out.Suggestions, report.SuggestLogOptsDaemon, report.SuggestLogOptsContainer)
	}
	return out, warns
}

// measure sums every regular file in dir whose name is base plus a rotation
// suffix. Any error leaves the item unmeasured and the section incomplete.
func measure(it *report.LogItem, dir, base string, out *report.Logs, warns *[]string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		it.Note = describeFSError(err)
		out.Complete = false
		*warns = append(*warns, fmt.Sprintf("logs: %s not measured: %s", it.Name, it.Note))
		return
	}
	// A file rotated away between ReadDir and Lstat is skipped: a slight
	// undercount for that instant, never an error.
	var bytes int64
	files := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, base) || !rotatedSuffix.MatchString(strings.TrimPrefix(name, base)) {
			continue
		}
		// Lstat: a symlink is not a log file and is never followed.
		info, err := os.Lstat(filepath.Join(dir, name))
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		bytes += info.Size()
		files++
	}
	it.Bytes = bytes
	it.Files = files
	it.Measured = true
	out.TotalBytes += bytes
}

// underRoot reports whether path stays inside root after cleaning, so a
// crafted LogPath cannot point dockdu at an arbitrary directory.
func underRoot(path, root string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func rotationJSONFile(opts map[string]string) (rotation, maxSize, maxFile string) {
	if ms := opts["max-size"]; ms != "" {
		mf := opts["max-file"]
		if mf == "" {
			mf = "1 (driver default)"
		}
		return "bounded", ms, mf
	}
	return "unbounded", "", ""
}

func optOrDefault(opts map[string]string, key, def string) string {
	if v := opts[key]; v != "" {
		return v
	}
	return def
}

func describeFSError(err error) string {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "log directory not found under the data root (mount /var/lib/docker read-only or run on the host)"
	case errors.Is(err, fs.ErrPermission):
		return "log directory not readable (run as root or inside the container image)"
	default:
		return err.Error()
	}
}
