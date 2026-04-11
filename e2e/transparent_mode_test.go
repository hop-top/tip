package e2e_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTransparentModeNoOutput(t *testing.T) {
	stateDir := freshStateDir(t)
	adapterDir := mockAdapterDir(t, map[string]map[string]any{
		"git": {
			"suggestions": []any{
				map[string]any{
					"text":       "git push",
					"confidence": 0.9,
					"primary":    true,
				},
			},
		},
	})

	// Set transparent mode
	modeFile := filepath.Join(stateDir, "mode")
	if err := os.WriteFile(modeFile, []byte("transparent"), 0644); err != nil {
		t.Fatalf("write mode: %v", err)
	}

	env := map[string]string{
		"TIP_STATE_DIR":    stateDir,
		"TIP_ADAPTER_PATH": adapterDir,
	}

	stdout, _, code := runTip(t,
		[]string{"observe", "--cmd-line", "git commit -m fix",
			"--exit", "0"},
		env,
	)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("transparent mode should produce no output, got: %q",
			stdout)
	}
}

func TestTransparentModeRecordsSession(t *testing.T) {
	stateDir := freshStateDir(t)

	// Set transparent mode
	modeFile := filepath.Join(stateDir, "mode")
	if err := os.WriteFile(modeFile, []byte("transparent"), 0644); err != nil {
		t.Fatalf("write mode: %v", err)
	}

	env := map[string]string{
		"TIP_STATE_DIR":    stateDir,
		"TIP_ADAPTER_PATH": t.TempDir(),
	}

	runTip(t,
		[]string{"observe", "--cmd-line", "git status",
			"--exit", "0"},
		env,
	)

	// Verify session file was written
	files, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatalf("read state dir: %v", err)
	}
	var found bool
	for _, f := range files {
		if !strings.HasPrefix(f.Name(), "session-") {
			continue
		}
		data, err := os.ReadFile(
			filepath.Join(stateDir, f.Name()))
		if err != nil {
			continue
		}
		var entries []map[string]any
		if json.Unmarshal(data, &entries) == nil &&
			len(entries) > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("transparent mode should still record session")
	}
}
