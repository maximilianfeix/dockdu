package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/moby/moby/client"
)

// fakeEngine is an instrumented Docker Engine API server. It answers the
// four read endpoints dockdu uses and records every request it receives,
// including the ones it does not know, so tests can assert that nothing
// else ever reached the daemon.
type fakeEngine struct {
	t          testing.TB
	containers int
	latency    time.Duration  // added to every inspect, to model a slow daemon
	failID     map[string]int // container id -> HTTP status for its inspect
	mu         sync.Mutex
	seen       []string            // "METHOD /path"
	inspected  map[string]struct{} // container ids inspected
}

func (f *fakeEngine) id(i int) string { return fmt.Sprintf("%064x", i+1) }

var inspectPath = regexp.MustCompile(`^/v[0-9.]+/containers/([0-9a-f]+)/json$`)

func (f *fakeEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.seen = append(f.seen, r.Method+" "+r.URL.Path)
	f.mu.Unlock()
	w.Header().Set("Api-Version", "1.53")
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, `{"message":"fake engine: mutating request"}`, http.StatusMethodNotAllowed)
		return
	}
	switch p := r.URL.Path; {
	case p == "/_ping":
		_, _ = w.Write([]byte("OK"))
	case strings.HasSuffix(p, "/info"):
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ServerVersion": "29.2.0", "OperatingSystem": "fake", "Driver": "overlay2",
			"LoggingDriver": "json-file", "DockerRootDir": "/var/lib/docker",
		})
	case strings.HasSuffix(p, "/system/df"):
		items := make([]map[string]any, f.containers)
		for i := range items {
			items[i] = map[string]any{"Id": f.id(i), "Names": []string{fmt.Sprintf("/c%04d", i)}, "State": "running", "SizeRw": 10}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ImageUsage":      map[string]any{"TotalSize": 100, "TotalCount": 1, "Items": []map[string]any{{"Id": "sha256:aa", "Size": 100, "SharedSize": 0, "Containers": f.containers}}},
			"ContainerUsage":  map[string]any{"TotalSize": 10 * f.containers, "TotalCount": f.containers, "Items": items},
			"VolumeUsage":     map[string]any{"Items": []any{}},
			"BuildCacheUsage": map[string]any{"Items": []any{}},
		})
	case inspectPath.MatchString(p):
		id := inspectPath.FindStringSubmatch(p)[1]
		if f.latency > 0 {
			time.Sleep(f.latency)
		}
		f.mu.Lock()
		if f.inspected == nil {
			f.inspected = map[string]struct{}{}
		}
		f.inspected[id] = struct{}{}
		status := f.failID[id]
		f.mu.Unlock()
		if status != 0 {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"message":"No such container: ` + id + `"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"Id": id, "LogPath": "/var/lib/docker/containers/" + id + "/" + id + "-json.log",
			"HostConfig": map[string]any{"LogConfig": map[string]any{"Type": "json-file", "Config": map[string]string{"max-size": "10m"}}},
		})
	default:
		f.t.Errorf("fake engine: unexpected request %s %s", r.Method, p)
		http.Error(w, `{"message":"unexpected"}`, http.StatusNotFound)
	}
}

func (f *fakeEngine) requests() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.seen...)
}

// newFake starts the fake engine and a Moby engine connected to it. The
// DOCKER_* variables are cleared so the developer's environment cannot
// redirect the test to a real daemon.
func newFake(t testing.TB, f *fakeEngine) *Moby {
	for _, k := range []string{"DOCKER_HOST", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH", "DOCKER_API_VERSION", "DOCKER_CONTEXT"} {
		t.Setenv(k, "")
	}
	f.t = t
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	m, err := NewMoby(client.WithHost("tcp://" + srv.Listener.Addr().String()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Close() })
	return m
}

var allowedRequest = regexp.MustCompile(`^(GET|HEAD) (/_ping|/v1\.53/(info|system/df|containers/[0-9a-f]{64}/json))$`)

func TestSnapshotUsesOnlyReadOnlyEndpoints(t *testing.T) {
	f := &fakeEngine{containers: 25}
	m := newFake(t, f)
	s, err := m.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range f.requests() {
		if !allowedRequest.MatchString(r) {
			t.Errorf("request outside the read-only allowlist: %s", r)
		}
	}
	if len(f.inspected) != 25 {
		t.Fatalf("inspected %d containers, want 25", len(f.inspected))
	}
	// Order follows the daemon's list even though inspects run concurrently.
	for i, c := range s.Containers {
		if c.ID != f.id(i) || c.LogDriver != "json-file" || c.LogOpts["max-size"] != "10m" {
			t.Fatalf("container %d = %+v", i, c)
		}
	}
	if len(s.Warnings) != 0 {
		t.Fatalf("unexpected warnings %v", s.Warnings)
	}
}

// TestGuardRefusesMutatingCalls is the regression test for the read-only
// contract: every state-changing client call must fail inside dockdu
// before a request reaches the daemon.
func TestGuardRefusesMutatingCalls(t *testing.T) {
	f := &fakeEngine{}
	m := newFake(t, f)
	ctx := context.Background()
	id := f.id(0)
	calls := map[string]func() error{
		"ContainerRemove": func() error { _, err := m.cli.ContainerRemove(ctx, id, client.ContainerRemoveOptions{}); return err },
		"ContainerStop":   func() error { _, err := m.cli.ContainerStop(ctx, id, client.ContainerStopOptions{}); return err },
		"ContainerKill":   func() error { _, err := m.cli.ContainerKill(ctx, id, client.ContainerKillOptions{}); return err },
		"ContainerCreate": func() error {
			_, err := m.cli.ContainerCreate(ctx, client.ContainerCreateOptions{Image: "alpine"})
			return err
		},
		"ExecCreate":      func() error { _, err := m.cli.ExecCreate(ctx, id, client.ExecCreateOptions{}); return err },
		"ImageRemove":     func() error { _, err := m.cli.ImageRemove(ctx, "sha256:aa", client.ImageRemoveOptions{}); return err },
		"VolumeRemove":    func() error { _, err := m.cli.VolumeRemove(ctx, "data", client.VolumeRemoveOptions{}); return err },
		"ContainerPrune":  func() error { _, err := m.cli.ContainerPrune(ctx, client.ContainerPruneOptions{}); return err },
		"ImagePrune":      func() error { _, err := m.cli.ImagePrune(ctx, client.ImagePruneOptions{}); return err },
		"VolumePrune":     func() error { _, err := m.cli.VolumePrune(ctx, client.VolumePruneOptions{}); return err },
		"BuildCachePrune": func() error { _, err := m.cli.BuildCachePrune(ctx, client.BuildCachePruneOptions{}); return err },
		// Reads dockdu does not need are refused too: logs and file archives
		// would expose container contents.
		"ContainerLogs": func() error { _, err := m.cli.ContainerLogs(ctx, id, client.ContainerLogsOptions{}); return err },
		"CopyFromContainer": func() error {
			_, err := m.cli.CopyFromContainer(ctx, id, client.CopyFromContainerOptions{SourcePath: "/etc/passwd"})
			return err
		},
	}
	for name, call := range calls {
		if err := call(); !errors.Is(err, ErrNotReadOnly) {
			t.Errorf("%s: err = %v, want ErrNotReadOnly", name, err)
		}
	}
	for _, r := range f.requests() {
		if !allowedRequest.MatchString(r) {
			t.Errorf("a refused call still reached the daemon: %s", r)
		}
	}
}

func TestReadOnlyGuardTable(t *testing.T) {
	id := strings.Repeat("ab", 32)
	cases := []struct {
		method, path string
		ok           bool
	}{
		{"HEAD", "/_ping", true},
		{"GET", "/_ping", true},
		{"GET", "/v1.53/info", true},
		{"GET", "/v1.44/system/df", true},
		{"GET", "/v1.53/containers/" + id + "/json", true},
		{"POST", "/v1.53/info", false},
		{"DELETE", "/v1.53/containers/" + id, false},
		{"GET", "/v1.53/containers/" + id + "/logs", false},
		{"GET", "/v1.53/containers/" + id + "/archive", false},
		{"GET", "/v1.53/containers/../json", false},
		{"GET", "/v1.53/containers/json", false},
		{"GET", "/v1.53/images/json", false},
		{"GET", "/v1.53/info/../containers/x/exec", false},
	}
	for _, c := range cases {
		req, err := http.NewRequest(c.method, "http://docker"+c.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got := readOnlyGuard(req) == nil; got != c.ok {
			t.Errorf("%s %s: allowed = %v, want %v", c.method, c.path, got, c.ok)
		}
	}
}

func TestSnapshotInspectFailureIsPartialNotFatal(t *testing.T) {
	f := &fakeEngine{containers: 3}
	f.failID = map[string]int{f.id(1): http.StatusNotFound} // removed while dockdu ran
	m := newFake(t, f)
	s, err := m.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if s.Containers[1].InspectError == "" || s.Containers[0].InspectError != "" || s.Containers[2].InspectError != "" {
		t.Fatalf("only container 1 should carry an inspect error: %+v", s.Containers)
	}
	if len(s.Warnings) != 1 || !strings.Contains(s.Warnings[0], "c0001") {
		t.Fatalf("want one warning naming c0001, got %v", s.Warnings)
	}
}

func TestSnapshotHonoursCancellation(t *testing.T) {
	f := &fakeEngine{containers: 50, latency: 20 * time.Millisecond}
	m := newFake(t, f)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := m.Snapshot(ctx); err == nil {
		t.Fatal("a cancelled snapshot must fail, not return partial data as complete")
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("cancellation took %v", d)
	}
}

// BenchmarkSnapshotSlowInspect models a host with 200 containers and a
// daemon that needs 2 ms per inspect; workers=1 is the v0.1 behaviour.
func BenchmarkSnapshotSlowInspect(b *testing.B) {
	for _, workers := range []int{1, 4} {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			defer func(w int) { inspectWorkers = w }(inspectWorkers)
			inspectWorkers = workers
			f := &fakeEngine{containers: 200, latency: 2 * time.Millisecond}
			m := newFake(b, f)
			for b.Loop() {
				if _, err := m.Snapshot(context.Background()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
