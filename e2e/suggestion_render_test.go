package e2e_test

import (
	"strings"
	"testing"
)

func TestObserveWithAdapter(t *testing.T) {
	stateDir := freshStateDir(t)
	adapterDir := mockAdapterDir(t, map[string]map[string]any{
		"git": {
			"suggestions": []any{
				map[string]any{
					"text":       "git push",
					"confidence": 0.9,
					"primary":    true,
				},
				map[string]any{
					"text":       "git log --oneline -5",
					"confidence": 0.5,
				},
			},
		},
	})

	env := map[string]string{
		"TIP_STATE_DIR":    stateDir,
		"TIP_ADAPTER_PATH": adapterDir,
	}
	stdout, _, code := runTip(t,
		[]string{"observe", "--cmd-line", "git commit -m fix", "--exit", "0"},
		env,
	)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}
	if !strings.Contains(stdout, "git push") {
		t.Errorf("missing primary suggestion, got: %q", stdout)
	}
	if !strings.Contains(stdout, "git log") {
		t.Errorf("missing secondary suggestion, got: %q", stdout)
	}
	// Should contain dimmed ANSI
	if !strings.Contains(stdout, "\033[2m") {
		t.Error("expected ANSI dim codes in output")
	}
}

func TestObserveNoAdapter(t *testing.T) {
	stateDir := freshStateDir(t)
	env := map[string]string{
		"TIP_STATE_DIR":    stateDir,
		"TIP_ADAPTER_PATH": t.TempDir(), // empty dir
	}
	stdout, _, code := runTip(t,
		[]string{"observe", "--cmd-line", "foo bar", "--exit", "0"},
		env,
	)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}
	// No adapter = no output
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("expected empty output, got: %q", stdout)
	}
}

func TestObserveNoColor(t *testing.T) {
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

	env := map[string]string{
		"TIP_STATE_DIR":    stateDir,
		"TIP_ADAPTER_PATH": adapterDir,
		"NO_COLOR":         "1",
	}
	stdout, _, code := runTip(t,
		[]string{"observe", "--cmd-line", "git commit -m fix", "--exit", "0"},
		env,
	)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}
	if strings.Contains(stdout, "\033[") {
		t.Error("NO_COLOR set but output contains ANSI codes")
	}
	if !strings.Contains(stdout, "git push") {
		t.Error("missing suggestion text")
	}
}
