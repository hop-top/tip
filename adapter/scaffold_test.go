package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	outDir, err := Scaffold(dir, "mytools")
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	wantDir := filepath.Join(dir, "tip-mytools")
	if outDir != wantDir {
		t.Errorf("outDir = %q, want %q", outDir, wantDir)
	}

	// Check expected files exist.
	for _, name := range []string{"go.mod", "main.go", "main_test.go"} {
		p := filepath.Join(outDir, name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", name)
		}
	}
}

func TestScaffoldGoModContent(t *testing.T) {
	dir := t.TempDir()
	outDir, err := Scaffold(dir, "docker")
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "module hop.top/tip-docker") {
		t.Errorf("go.mod missing module line, got:\n%s", content)
	}
	if !strings.Contains(content, "go 1.26") {
		t.Errorf("go.mod missing go version, got:\n%s", content)
	}
}

func TestScaffoldMainGoContent(t *testing.T) {
	dir := t.TempDir()
	outDir, err := Scaffold(dir, "gh")
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, `"gh"`) {
		t.Error("main.go should contain adapter name in describe")
	}
	if !strings.Contains(content, "tip-gh") {
		t.Error("main.go should reference tip-gh in usage")
	}
}

func TestScaffoldTestContent(t *testing.T) {
	dir := t.TempDir()
	outDir, err := Scaffold(dir, "wrangler")
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "main_test.go"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "wrangler") {
		t.Error("main_test.go should contain adapter name")
	}
}
