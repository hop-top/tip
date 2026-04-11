package adapter

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDiscoverInPATH(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script not executable on Windows")
	}

	tmp := t.TempDir()
	bin := filepath.Join(tmp, "tip-fakecli")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", tmp+string(os.PathListSeparator)+os.Getenv("PATH"))
	got := Discover("fakecli")
	if got == "" {
		t.Fatal("expected to find tip-fakecli in PATH")
	}
	if filepath.Base(got) != "tip-fakecli" {
		t.Errorf("got %q, want tip-fakecli", filepath.Base(got))
	}
}

func TestDiscoverInPluginDir(t *testing.T) {
	home := t.TempDir()
	plugDir := filepath.Join(home, PluginDir)
	if err := os.MkdirAll(plugDir, 0o755); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(plugDir, "tip-myplugin")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Ensure it's NOT in $PATH
	t.Setenv("PATH", t.TempDir())
	got := DiscoverWithHome("myplugin", home)
	if got == "" {
		t.Fatal("expected to find tip-myplugin in plugin dir")
	}
	if got != bin {
		t.Errorf("got %q, want %q", got, bin)
	}
}

func TestDiscoverNotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	got := DiscoverWithHome("nonexistent", t.TempDir())
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
