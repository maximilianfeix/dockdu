package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/moby/client"
)

// Target is the daemon endpoint dockdu connects to and where that choice
// came from, so an unreachable daemon can be explained precisely.
type Target struct {
	// Source is "DOCKER_HOST", `context "name"` or "default".
	Source string
	// Opts are passed to NewMoby; they are applied after client.FromEnv.
	Opts []client.Opt
}

// maxConfigBytes bounds the CLI files dockdu reads; real ones are < 10 KiB.
const maxConfigBytes = 1 << 20

// ResolveTarget picks the daemon the docker CLI would use, without running
// it: DOCKER_HOST wins, then DOCKER_CONTEXT, then currentContext from the
// CLI config file. Contexts are how Colima, OrbStack, rootless Docker and
// Docker Desktop on Linux point the CLI at a socket other than
// /var/run/docker.sock. Only unix://, npipe:// and tcp:// endpoints are
// supported; ssh:// needs the docker CLI as a connection helper, which
// dockdu does not run.
//
// getenv is os.Getenv in production; home is the user's home directory.
func ResolveTarget(getenv func(string) string, home string) (Target, error) {
	if getenv("DOCKER_HOST") != "" {
		return Target{Source: "DOCKER_HOST"}, nil
	}
	configDir := getenv("DOCKER_CONFIG")
	if configDir == "" {
		if home == "" {
			return Target{Source: "default"}, nil
		}
		configDir = filepath.Join(home, ".docker")
	}
	name := getenv("DOCKER_CONTEXT")
	if name == "" {
		path := filepath.Join(configDir, "config.json")
		cfg, err := readLimited(path)
		if errors.Is(err, fs.ErrNotExist) {
			return Target{Source: "default"}, nil
		}
		if err != nil {
			return Target{}, fmt.Errorf("read docker CLI config: %w", err)
		}
		var c struct {
			CurrentContext string `json:"currentContext"`
		}
		if err := json.Unmarshal(cfg, &c); err != nil {
			return Target{}, fmt.Errorf("parse docker CLI config %s: %w", path, err)
		}
		name = c.CurrentContext
	}
	if name == "" || name == "default" {
		return Target{Source: "default"}, nil
	}

	src := fmt.Sprintf("context %q", name)
	sum := sha256.Sum256([]byte(name))
	id := hex.EncodeToString(sum[:])
	raw, err := readLimited(filepath.Join(configDir, "contexts", "meta", id, "meta.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return Target{}, fmt.Errorf("docker %s not found (see `docker context ls`)", src)
	}
	if err != nil {
		return Target{}, fmt.Errorf("docker %s: read context metadata: %w", src, err)
	}
	var meta struct {
		Endpoints struct {
			Docker struct {
				Host          string `json:"Host"`
				SkipTLSVerify bool   `json:"SkipTLSVerify"`
			} `json:"docker"`
		} `json:"Endpoints"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return Target{}, fmt.Errorf("docker %s: parse context metadata: %w", src, err)
	}
	host := meta.Endpoints.Docker.Host
	scheme, _, _ := strings.Cut(host, "://")
	switch scheme {
	case "unix", "npipe":
		return Target{Source: src, Opts: []client.Opt{client.WithHost(host)}}, nil
	case "tcp":
		if meta.Endpoints.Docker.SkipTLSVerify {
			return Target{}, fmt.Errorf("docker %s: SkipTLSVerify is not supported; set DOCKER_HOST and DOCKER_CERT_PATH instead", src)
		}
		opts := []client.Opt{client.WithHost(host)}
		tlsDir := filepath.Join(configDir, "contexts", "tls", id, "docker")
		_, err := os.Stat(tlsDir)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			// Unreadable TLS material must not silently become plain HTTP.
			return Target{}, fmt.Errorf("docker %s: TLS files: %w", src, err)
		}
		if err == nil {
			opts = append(opts, client.WithTLSClientConfig(
				existing(filepath.Join(tlsDir, "ca.pem")),
				existing(filepath.Join(tlsDir, "cert.pem")),
				existing(filepath.Join(tlsDir, "key.pem"))))
		}
		return Target{Source: src, Opts: opts}, nil
	case "":
		return Target{}, fmt.Errorf("docker %s has no docker endpoint", src)
	default:
		return Target{}, fmt.Errorf("docker %s uses %s://, which dockdu does not support; set DOCKER_HOST to a unix:// or tcp:// address", src, scheme)
	}
}

// readLimited reads a small config file and refuses anything larger than
// maxConfigBytes instead of loading it.
func readLimited(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, maxConfigBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > maxConfigBytes {
		return nil, fmt.Errorf("%s is larger than %d bytes", path, maxConfigBytes)
	}
	return raw, nil
}

// existing returns path when the file exists and "" otherwise, so absent
// TLS files are left out instead of failing the client setup.
func existing(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}
