package e2e_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionBuildsHistory(t *testing.T) {
	stateDir := freshStateDir(t)
	env := map[string]string{
		"TIP_STATE_DIR":    stateDir,
		"TIP_ADAPTER_PATH": t.TempDir(),
	}

	// Multiple observe calls
	cmds := []string{
		"git add .",
		"git commit -m fix",
		"git push",
	}
	for _, c := range cmds {
		_, _, code := runTip(t,
			[]string{"observe", "--cmd-line", c, "--exit", "0"},
			env,
		)
		if code != 0 {
			t.Fatalf("observe %q exit code: %d", c, code)
		}
	}

	// Verify session file exists with entries
	files, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatalf("read state dir: %v", err)
	}
	var sessionFile string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "session-") {
			sessionFile = filepath.Join(stateDir, f.Name())
			break
		}
	}
	if sessionFile == "" {
		t.Fatal("no session file created")
	}

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("read session: %v", err)
	}

	var entries []map[string]any
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("parse session: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0]["cmd"] != "git" {
		t.Errorf("entry 0 cmd: %v", entries[0]["cmd"])
	}
}

func TestSessionScopedToCwd(t *testing.T) {
	stateDir := freshStateDir(t)
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	env := map[string]string{
		"TIP_STATE_DIR":    stateDir,
		"TIP_ADAPTER_PATH": t.TempDir(),
	}

	// Observe in dir1
	_, _, code1 := runTipInDir(t, dir1,
		[]string{"observe", "--cmd-line", "ls -la", "--exit", "0"},
		env,
	)
	if code1 != 0 {
		t.Fatalf("dir1 observe exit code: %d", code1)
	}
	// Observe in dir2
	_, _, code2 := runTipInDir(t, dir2,
		[]string{"observe", "--cmd-line", "pwd", "--exit", "0"},
		env,
	)
	if code2 != 0 {
		t.Fatalf("dir2 observe exit code: %d", code2)
	}

	// Count session files — should be 2 different ones
	files, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatalf("read state dir: %v", err)
	}
	sessionCount := 0
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "session-") {
			sessionCount++
		}
	}
	if sessionCount < 2 {
		t.Errorf("expected 2+ session files for different dirs, got %d",
			sessionCount)
	}
}
