package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeContext lays out a docker CLI config dir the way `docker context
// create` does: config.json plus contexts/meta/<sha256(name)>/meta.json.
func writeContext(t *testing.T, dir, current, name, host string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if current != "" {
		if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"auths":{},"currentContext":"`+current+`"}`), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	sum := sha256.Sum256([]byte(name))
	meta := filepath.Join(dir, "contexts", "meta", hex.EncodeToString(sum[:]))
	if err := os.MkdirAll(meta, 0o700); err != nil {
		t.Fatal(err)
	}
	body := `{"Name":"` + name + `","Metadata":{},"Endpoints":{"docker":{"Host":"` + host + `","SkipTLSVerify":false}}}`
	if err := os.WriteFile(filepath.Join(meta, "meta.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestResolveTargetPrecedence(t *testing.T) {
	home := t.TempDir()
	cfg := filepath.Join(home, ".docker")
	writeContext(t, cfg, "colima", "colima", "unix:///Users/x/.colima/default/docker.sock")
	writeContext(t, cfg, "", "other", "unix:///run/other.sock")

	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"DOCKER_HOST wins", map[string]string{"DOCKER_HOST": "unix:///x.sock", "DOCKER_CONTEXT": "other"}, "DOCKER_HOST"},
		{"DOCKER_CONTEXT beats config", map[string]string{"DOCKER_CONTEXT": "other"}, `context "other"`},
		{"currentContext from config", nil, `context "colima"`},
		{"explicit default", map[string]string{"DOCKER_CONTEXT": "default"}, "default"},
	}
	for _, c := range cases {
		got, err := ResolveTarget(env(c.env), home)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got.Source != c.want {
			t.Errorf("%s: source = %q, want %q", c.name, got.Source, c.want)
		}
	}
}

func TestResolveTargetNoConfigIsDefault(t *testing.T) {
	got, err := ResolveTarget(env(nil), t.TempDir())
	if err != nil || got.Source != "default" || len(got.Opts) != 0 {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestResolveTargetDockerConfigDir(t *testing.T) {
	dir := t.TempDir()
	writeContext(t, dir, "rootless", "rootless", "unix:///run/user/1000/docker.sock")
	got, err := ResolveTarget(env(map[string]string{"DOCKER_CONFIG": dir}), "/nonexistent")
	if err != nil || got.Source != `context "rootless"` || len(got.Opts) != 1 {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestResolveTargetRejectsUnsupportedEndpoints(t *testing.T) {
	home := t.TempDir()
	cfg := filepath.Join(home, ".docker")
	writeContext(t, cfg, "remote", "remote", "ssh://user@host")
	_, err := ResolveTarget(env(nil), home)
	if err == nil || !strings.Contains(err.Error(), "ssh://") || !strings.Contains(err.Error(), "DOCKER_HOST") {
		t.Fatalf("ssh context must fail with a hint, got %v", err)
	}
}

func TestResolveTargetMissingContextIsAnError(t *testing.T) {
	home := t.TempDir()
	cfg := filepath.Join(home, ".docker")
	if err := os.MkdirAll(cfg, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg, "config.json"), []byte(`{"currentContext":"gone"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveTarget(env(nil), home); err == nil || !strings.Contains(err.Error(), `context "gone"`) {
		t.Fatalf("missing context metadata must be reported, got %v", err)
	}
}

func TestResolveTargetRefusesHugeConfig(t *testing.T) {
	home := t.TempDir()
	cfg := filepath.Join(home, ".docker")
	if err := os.MkdirAll(cfg, 0o700); err != nil {
		t.Fatal(err)
	}
	big := make([]byte, maxConfigBytes+10)
	if err := os.WriteFile(filepath.Join(cfg, "config.json"), big, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveTarget(env(nil), home); err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Fatalf("oversized config must be refused, got %v", err)
	}
}

func TestResolveTargetUnreadableTLSDirIsAnError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read the directory anyway")
	}
	home := t.TempDir()
	cfg := filepath.Join(home, ".docker")
	writeContext(t, cfg, "tls", "tls", "tcp://10.0.0.1:2376")
	sum := sha256.Sum256([]byte("tls"))
	parent := filepath.Join(cfg, "contexts", "tls", hex.EncodeToString(sum[:]))
	if err := os.MkdirAll(filepath.Join(parent, "docker"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(parent, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o700) })
	if _, err := ResolveTarget(env(nil), home); err == nil || !strings.Contains(err.Error(), "TLS files") {
		t.Fatalf("an unreadable TLS dir must not fall back to plain HTTP, got %v", err)
	}
}
