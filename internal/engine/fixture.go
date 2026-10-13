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

// WriteFixture stores a sanitised snapshot as indented JSON.
func WriteFixture(path string, s *Snapshot, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%w: %s", ErrFixtureExists, path)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
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
// namespace, the daemon host and name, and build-cache descriptions, which
// can contain local paths and secrets from Dockerfile instructions.
func Sanitize(s *Snapshot) {
	s.Daemon.Host = "unix:///var/run/docker.sock"
	for i := range s.Images {
		s.Images[i].Labels = filterLabels(s.Images[i].Labels)
	}
	for i := range s.Containers {
		s.Containers[i].Labels = filterLabels(s.Containers[i].Labels)
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

// redactDescription keeps only the instruction keyword (RUN, COPY, ...) of a
// build step description.
func redactDescription(d string) string {
	d = strings.TrimSpace(d)
	if d == "" {
		return ""
	}
	fields := strings.Fields(d)
	if len(fields) == 0 {
		return ""
	}
	if strings.HasPrefix(fields[0], "[") && len(fields) > 1 {
		return fields[0] + " " + fields[1] + " …"
	}
	return fields[0] + " …"
}
