package e2e_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// tipBin holds the path to the compiled tip binary.
var tipBin string

func TestMain(m *testing.M) {
	bin, err := buildTip()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build tip: %v\n", err)
		os.Exit(1)
	}
	tipBin = bin
	os.Exit(m.Run())
}

// buildTip compiles cmd/tip into a temp directory, returns path.
func buildTip() (string, error) {
	dir, err := os.MkdirTemp("", "tip-e2e-*")
	if err != nil {
		return "", err
	}
	bin := filepath.Join(dir, "tip")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	// Resolve project root (e2e/ is one level below)
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(thisFile))

	cmd := exec.Command("go", "build", "-tags", "e2e", "-o", bin, "./cmd/tip/")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%s: %w", string(out), err)
	}
	return bin, nil
}

// runTip executes the tip binary with args and env overrides.
// Returns stdout, stderr, exit code.
// dedupEnv returns os.Environ() with keys in overrides removed,
// then appends the overrides.
func dedupEnv(overrides map[string]string) []string {
	base := os.Environ()
	out := make([]string, 0, len(base)+len(overrides))
	for _, entry := range base {
		k, _, _ := strings.Cut(entry, "=")
		if _, dup := overrides[k]; dup {
			continue
		}
		out = append(out, entry)
	}
	for k, v := range overrides {
		out = append(out, k+"="+v)
	}
	return out
}

func runTip(
	t *testing.T,
	args []string,
	env map[string]string,
) (string, string, int) {
	t.Helper()
	cmd := exec.Command(tipBin, args...)
	cmd.Env = dedupEnv(env)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			t.Fatalf("exec error: %v", err)
		}
	}
	return stdout.String(), stderr.String(), code
}

// runTipInDir runs tip in a specific working directory.
func runTipInDir(
	t *testing.T,
	dir string,
	args []string,
	env map[string]string,
) (string, string, int) {
	t.Helper()
	cmd := exec.Command(tipBin, args...)
	cmd.Dir = dir
	cmd.Env = dedupEnv(env)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			t.Fatalf("exec error: %v", err)
		}
	}
	return stdout.String(), stderr.String(), code
}

// mockAdapterDir creates a temp dir with mock adapter binaries.
// Each adapter is a tiny Go program that reads stdin JSON and
// writes a configured response to stdout.
func mockAdapterDir(
	t *testing.T,
	adapters map[string]map[string]any,
) string {
	t.Helper()
	dir := t.TempDir()
	for name, response := range adapters {
		buildMockAdapter(t, dir, name, response)
	}
	return dir
}

// buildMockAdapter compiles a mock adapter binary that returns
// the given JSON response on "suggest" subcommand.
func buildMockAdapter(
	t *testing.T,
	dir, name string,
	response map[string]any,
) {
	t.Helper()
	respJSON, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal mock response: %v", err)
	}

	descJSON := fmt.Sprintf(
		`{"name":%q,"version":"0.0.1"}`, name)

	src := fmt.Sprintf(`package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}
	switch os.Args[1] {
	case "suggest":
		fmt.Print(%s)
	case "describe":
		fmt.Print(%s)
	default:
		os.Exit(1)
	}
}
`, "`"+string(respJSON)+"`", "`"+descJSON+"`")

	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "main.go")
	if err := os.WriteFile(srcFile, []byte(src), 0644); err != nil {
		t.Fatalf("write mock source: %v", err)
	}

	bin := filepath.Join(dir, "tip-"+name)
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, srcFile)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build mock adapter %s: %s: %v",
			name, string(out), err)
	}
}

// freshStateDir returns a temp dir for TIP_STATE_DIR isolation.
func freshStateDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func TestBuildTip(t *testing.T) {
	if tipBin == "" {
		t.Fatal("tip binary not built")
	}
	fi, err := os.Stat(tipBin)
	if err != nil {
		t.Fatalf("stat tip binary: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatal("tip binary is empty")
	}
}

func TestMockAdapterBuilds(t *testing.T) {
	resp := map[string]any{
		"suggestions": []any{
			map[string]any{
				"text":       "echo hello",
				"confidence": 0.9,
				"primary":    true,
			},
		},
	}
	dir := mockAdapterDir(t, map[string]map[string]any{
		"test": resp,
	})
	bin := filepath.Join(dir, "tip-test")
	fi, err := os.Stat(bin)
	if err != nil {
		t.Fatalf("stat mock adapter: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatal("mock adapter is empty")
	}

	// Verify it runs
	cmd := exec.Command(bin, "suggest")
	cmd.Stdin = strings.NewReader(`{"cmd":"test"}`)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("run mock adapter: %v", err)
	}
	if !strings.Contains(string(out), "echo hello") {
		t.Fatalf("unexpected output: %s", out)
	}
}
