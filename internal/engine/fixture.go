package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// Static is an Engine that returns a fixed Snapshot. It backs unit tests.
type Static struct{ S *Snapshot }

// Snapshot returns a deep copy so collectors cannot mutate the fixture.
func (st Static) Snapshot(context.Context) (*Snapshot, error) {
	if st.S == nil {
		return nil, errors.New("static engine: nil snapshot")
	}
	raw, err := json.Marshal(st.S)
	if err != nil {
		return nil, err
	}
	var out Snapshot
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// LoadFixture reads a Snapshot written by WriteFixture.
func LoadFixture(path string) (*Snapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read fixture: %w", err)
	}
	var s Snapshot
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("parse fixture %s: %w", path, err)
	}
	return &s, nil
}

// ErrFixtureExists is returned by WriteFixture when the target exists and
// overwrite is false. Recording never silently replaces test data.
var ErrFixtureExists = errors.New("fixture exists; pass overwrite to replace it")

// WriteFixture stores a sanitised snapshot as indented JSON. Without
// overwrite the file is created with O_EXCL, so an existing file or a
// symlink planted at path is never replaced (no stat-then-write race).
func WriteFixture(path string, s *Snapshot, overwrite bool) error {
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if !overwrite {
		flags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	}
	f, err := os.OpenFile(path, flags, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("%w: %s", ErrFixtureExists, path)
	}
	if err != nil {
		return err
	}
	if _, err := f.Write(append(raw, '\n')); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// Record takes a snapshot from e, sanitises it and writes it to path.
func Record(ctx context.Context, e Engine, path string, overwrite bool) error {
	s, err := e.Snapshot(ctx)
	if err != nil {
		return err
	}
	Sanitize(s)
	return WriteFixture(path, s, overwrite)
}

// keptLabelPrefixes are the only label namespaces a fixture keeps. Everything
// else may carry private data (build args, hostnames, tickets) and is dropped.
var keptLabelPrefixes = []string{"com.docker.compose.", "dockdu."}

// Sanitize removes data a fixture does not need: labels outside the compose
// namespace, the daemon host and name, registry hosts in image references,
// and build-cache descriptions, which can contain local paths and secrets
// from Dockerfile instructions.
func Sanitize(s *Snapshot) {
	s.Daemon.Host = "unix:///var/run/docker.sock"
	for i := range s.Images {
		s.Images[i].Labels = filterLabels(s.Images[i].Labels)
		for j, tag := range s.Images[i].RepoTags {
			s.Images[i].RepoTags[j] = redactRegistry(tag)
		}
	}
	for i := range s.Containers {
		s.Containers[i].Labels = filterLabels(s.Containers[i].Labels)
		s.Containers[i].Image = redactRegistry(s.Containers[i].Image)
		if wd, ok := s.Containers[i].Labels["com.docker.compose.project.working_dir"]; ok {
			s.Containers[i].Labels["com.docker.compose.project.working_dir"] = redactPath(wd)
		}
		if cf, ok := s.Containers[i].Labels["com.docker.compose.project.config_files"]; ok {
			s.Containers[i].Labels["com.docker.compose.project.config_files"] = redactPath(cf)
		}
	}
	for i := range s.Volumes {
		s.Volumes[i].Labels = filterLabels(s.Volumes[i].Labels)
	}
	for i := range s.BuildCache {
		s.BuildCache[i].Description = redactDescription(s.BuildCache[i].Description)
	}
}

func filterLabels(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := map[string]string{}
	for k, v := range m {
		for _, p := range keptLabelPrefixes {
			if strings.HasPrefix(k, p) {
				out[k] = v
				break
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// redactPath keeps the last path element so grouping stays recognisable.
func redactPath(p string) string {
	parts := strings.Split(p, ",")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if idx := strings.LastIndex(part, "/"); idx >= 0 {
			parts[i] = "<redacted>/" + part[idx+1:]
		} else {
			parts[i] = part
		}
	}
	return strings.Join(parts, ",")
}

// redactRegistry replaces a registry host in an image reference with
// <registry>, keeping the repository path and tag so grouping stays
// readable. Docker treats the first path component as a registry when it
// contains a dot or a colon or is "localhost".
func redactRegistry(ref string) string {
	first, rest, ok := strings.Cut(ref, "/")
	if !ok {
		return ref
	}
	if strings.ContainsAny(first, ".:") || first == "localhost" {
		return "<registry>/" + rest
	}
	return ref
}

// redactDescription keeps only the stage marker and the instruction keyword
// (RUN, COPY, ...) of a build step description.
func redactDescription(d string) string {
	fields := strings.Fields(d)
	if len(fields) == 0 {
		return ""
	}
	i := 0
	if strings.HasPrefix(fields[0], "[") {
		// "[3/5] RUN …" or "[stage-1 4/4] RUN …": skip through the closing bracket.
		for i < len(fields) && !strings.HasSuffix(fields[i], "]") {
			i++
		}
		i++
	}
	marker := strings.Join(fields[:min(i, len(fields))], " ")
	if i >= len(fields) {
		return strings.TrimSpace(marker + " …")
	}
	return strings.TrimSpace(marker + " " + fields[i] + " …")
}
