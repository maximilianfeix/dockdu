package engine

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func sample() *Snapshot {
	last := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	return &Snapshot{
		TakenAt: last,
		Daemon:  Daemon{Host: "unix:///Users/someone/.docker/run/docker.sock", ServerVersion: "29.2.0", DataRoot: "/var/lib/docker"},
		Totals:  Totals{Images: Aggregate{1, 2, 300, 1000}},
		Images: []Image{{ID: "sha256:aaa", RepoTags: []string{"alpine:3.22"}, Size: 1000, SharedSize: 200, Containers: 1,
			Labels: map[string]string{"maintainer": "someone@example.com", "com.docker.compose.project": "shop"}}},
		Containers: []Container{{ID: "c1", Name: "shop-web-1", LogDriver: "json-file",
			Labels: map[string]string{"com.docker.compose.project": "shop", "com.docker.compose.project.working_dir": "/Users/someone/src/shop", "ticket": "ABC-1"}}},
		Volumes:    []Volume{{Name: "shop_db", HasUsageData: true, Size: 10, RefCount: 1, Labels: map[string]string{"backup": "nightly"}}},
		BuildCache: []CacheRecord{{ID: "r1", Type: "regular", Description: "RUN /bin/sh -c npm ci --token=secret", Size: 5, CreatedAt: last, LastUsedAt: &last}},
	}
}

func TestSanitizeDropsPrivateData(t *testing.T) {
	s := sample()
	Sanitize(s)
	if s.Daemon.Host != "unix:///var/run/docker.sock" {
		t.Errorf("host not redacted: %q", s.Daemon.Host)
	}
	if _, ok := s.Images[0].Labels["maintainer"]; ok {
		t.Errorf("non-compose image label kept")
	}
	if s.Images[0].Labels["com.docker.compose.project"] != "shop" {
		t.Errorf("compose label dropped")
	}
	if _, ok := s.Containers[0].Labels["ticket"]; ok {
		t.Errorf("non-compose container label kept")
	}
	if got := s.Containers[0].Labels["com.docker.compose.project.working_dir"]; got != "<redacted>/shop" {
		t.Errorf("working_dir = %q", got)
	}
	if s.Volumes[0].Labels != nil {
		t.Errorf("volume labels should be nil, got %v", s.Volumes[0].Labels)
	}
	if got := s.BuildCache[0].Description; got != "RUN …" {
		t.Errorf("description = %q", got)
	}
}

func TestWriteLoadRoundTripAndNoOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fx.json")
	s := sample()
	if err := WriteFixture(path, s, false); err != nil {
		t.Fatal(err)
	}
	if err := WriteFixture(path, s, false); !errors.Is(err, ErrFixtureExists) {
		t.Fatalf("second write should refuse, got %v", err)
	}
	if err := WriteFixture(path, s, true); err != nil {
		t.Fatalf("overwrite should succeed, got %v", err)
	}
	got, err := LoadFixture(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Images[0].ID != "sha256:aaa" || got.Volumes[0].Size != 10 || got.BuildCache[0].LastUsedAt == nil {
		t.Fatalf("round trip lost data: %+v", got)
	}
}

func TestStaticReturnsCopy(t *testing.T) {
	st := Static{S: sample()}
	a, err := st.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	a.Images[0].Size = 1
	b, _ := st.Snapshot(context.Background())
	if b.Images[0].Size != 1000 {
		t.Fatalf("static snapshot was mutated through a copy")
	}
}
