package collect

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"github.com/maximilianfeix/dockdu/internal/engine"
)

// TestIntegrationLogsRotation is SPEC §9.4. It creates one throwaway
// container labelled dockdu.test=1 with a 1k×3 json-file rotation, writes
// more than 3 KiB of output, and asserts the measured log footprint. It is
// the only test that changes Docker state, and only on the container it
// created, which it removes by ID.
//
// The log files live under the daemon's data root. On Docker Desktop that
// is inside the VM, so run this test inside a container with
// /var/run/docker.sock and /var/lib/docker:ro mounted (see Makefile
// `integration`). Without a visible data root the test skips.
func TestIntegrationLogsRotation(t *testing.T) {
	if os.Getenv("DOCKDU_INTEGRATION") != "1" {
		t.Skip("set DOCKDU_INTEGRATION=1 to run against a live daemon")
	}
	eng, err := engine.NewMoby()
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if _, err := eng.Snapshot(ctx); err != nil {
		if errors.Is(err, engine.ErrDaemonUnreachable) {
			t.Skipf("daemon unreachable: %v", err)
		}
		t.Fatal(err)
	}

	// The test talks to the daemon directly for its own resource; the
	// engine package stays read-only. Image alpine:3.22 must be present.
	cli, err := client.New(client.FromEnv)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	script := `i=0; while [ $i -lt 120 ]; do echo "dockdu integration line $i padding padding padding padding"; i=$((i+1)); done; sleep 300`
	created, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image:  "alpine:3.22",
			Cmd:    []string{"sh", "-c", script},
			Labels: map[string]string{"dockdu.test": "1"},
		},
		HostConfig: &container.HostConfig{
			LogConfig: container.LogConfig{Type: "json-file", Config: map[string]string{"max-size": "1k", "max-file": "3"}},
		},
	})
	if err != nil {
		t.Fatalf("container create (is alpine:3.22 pulled?): %v", err)
	}
	id := created.ID
	t.Cleanup(func() {
		_, _ = cli.ContainerRemove(context.Background(), id, client.ContainerRemoveOptions{Force: true})
	})
	if _, err := cli.ContainerStart(ctx, id, client.ContainerStartOptions{}); err != nil {
		t.Fatalf("container start: %v", err)
	}
	time.Sleep(2 * time.Second) // let the driver write and rotate

	s, err := eng.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(s.Daemon.DataRoot, "containers")); err != nil {
		t.Skipf("data root %s not visible from here (%v); run inside a container with /var/lib/docker mounted", s.Daemon.DataRoot, err)
	}
	logs, warns := Logs(s, "")
	var found bool
	for _, it := range logs.Items {
		if it.ContainerID != id {
			continue
		}
		found = true
		if !it.Measured {
			t.Fatalf("container not measured: %+v (warnings %v)", it, warns)
		}
		if it.Rotation != "bounded" || it.MaxSize != "1k" || it.MaxFile != "3" {
			t.Fatalf("rotation not detected: %+v", it)
		}
		if it.Files < 2 || it.Bytes < 1500 {
			t.Fatalf("expected at least two files and >1500 bytes incl. rotated ones, got files=%d bytes=%d", it.Files, it.Bytes)
		}
		t.Logf("measured %d bytes in %d files for %s", it.Bytes, it.Files, id[:12])
	}
	if !found {
		t.Fatalf("test container %s missing from snapshot", id[:12])
	}
}
