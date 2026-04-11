package adapter

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func buildDescribeAdapter(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	src := `package main

import (
	"encoding/json"
	"os"
)

type Resp struct {
	Name         string   ` + "`" + `json:"name"` + "`" + `
	Version      string   ` + "`" + `json:"version"` + "`" + `
	Commands     []string ` + "`" + `json:"commands"` + "`" + `
	Capabilities []string ` + "`" + `json:"capabilities"` + "`" + `
}

func main() {
	if len(os.Args) < 2 || os.Args[1] != "describe" {
		os.Exit(1)
	}
	resp := Resp{
		Name:     "testadapter",
		Version:  "1.0.0",
		Commands: []string{"run", "build"},
	}
	json.NewEncoder(os.Stdout).Encode(resp)
}
`
	srcFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(srcFile, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	bin := filepath.Join(dir, "tip-testadapter")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", bin, srcFile)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

func TestCachedDescribeMiss(t *testing.T) {
	bin := buildDescribeAdapter(t)
	cache := NewDescribeCache()

	resp, err := cache.CachedDescribe(bin)
	if err != nil {
		t.Fatalf("CachedDescribe: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil describe response")
	}
	if resp.Name != "testadapter" {
		t.Errorf("Name = %q, want testadapter", resp.Name)
	}
	if cache.Len() != 1 {
		t.Errorf("cache len = %d, want 1", cache.Len())
	}
}

func TestCachedDescribeHit(t *testing.T) {
	bin := buildDescribeAdapter(t)
	cache := NewDescribeCache()

	// First call populates cache.
	resp1, err := cache.CachedDescribe(bin)
	if err != nil {
		t.Fatal(err)
	}

	// Delete the binary so second call MUST use cache.
	os.Remove(bin)

	resp2, err := cache.CachedDescribe(bin)
	if err != nil {
		t.Fatal(err)
	}
	if resp2 == nil {
		t.Fatal("expected cached response")
	}
	if resp1.Name != resp2.Name {
		t.Errorf("cached name mismatch: %q vs %q", resp1.Name, resp2.Name)
	}
}

func TestCachedDescribeTTLExpiry(t *testing.T) {
	bin := buildDescribeAdapter(t)
	cache := NewDescribeCacheWithTTL(1 * time.Second)

	now := time.Now()
	cache.SetNow(func() time.Time { return now })

	// Populate cache.
	_, err := cache.CachedDescribe(bin)
	if err != nil {
		t.Fatal(err)
	}

	// Delete binary.
	os.Remove(bin)

	// Advance time past TTL.
	cache.SetNow(func() time.Time { return now.Add(2 * time.Second) })

	// Should attempt to re-fetch (and get nil since binary is gone).
	resp, err := cache.CachedDescribe(bin)
	if err != nil {
		t.Fatal(err)
	}
	// CallDescribe returns nil on error, which gets cached as nil.
	if resp != nil {
		t.Error("expected nil after TTL expiry with missing binary")
	}
}
