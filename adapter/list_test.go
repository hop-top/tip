package adapter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListFindsInPATH(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"tip-git", "tip-docker", "notanadapter"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("PATH", dir)
	infos := ListWithHome(t.TempDir()) // empty home so plugin dir is empty
	if len(infos) != 2 {
		t.Fatalf("got %d adapters, want 2: %v", len(infos), infos)
	}

	names := map[string]bool{}
	for _, info := range infos {
		names[info.Name] = true
	}
	if !names["git"] || !names["docker"] {
		t.Errorf("expected git and docker, got %v", names)
	}
}

func TestListFindsInPluginDir(t *testing.T) {
	home := t.TempDir()
	plugDir := filepath.Join(home, PluginDir)
	if err := os.MkdirAll(plugDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plugDir, "tip-wrangler"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", t.TempDir()) // empty PATH
	infos := ListWithHome(home)
	if len(infos) != 1 {
		t.Fatalf("got %d adapters, want 1", len(infos))
	}
	if infos[0].Name != "wrangler" {
		t.Errorf("name = %q, want wrangler", infos[0].Name)
	}
}

func TestListDeduplicates(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tip-git"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	home := t.TempDir()
	plugDir := filepath.Join(home, PluginDir)
	if err := os.MkdirAll(plugDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Same adapter in plugin dir — should be deduped.
	if err := os.WriteFile(filepath.Join(plugDir, "tip-git"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", dir)
	infos := ListWithHome(home)
	if len(infos) != 1 {
		t.Fatalf("got %d, want 1 (deduped)", len(infos))
	}
}

func TestListEmpty(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	infos := ListWithHome(t.TempDir())
	if len(infos) != 0 {
		t.Errorf("expected empty, got %d", len(infos))
	}
}
