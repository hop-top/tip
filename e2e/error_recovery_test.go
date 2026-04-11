package e2e_test

import (
	"strings"
	"testing"
)

func TestObserveErrorWithRecovery(t *testing.T) {
	stateDir := freshStateDir(t)
	adapterDir := mockAdapterDir(t, map[string]map[string]any{
		"docker": {
			"suggestions": []any{},
			"recovery":    "docker login registry",
		},
	})

	env := map[string]string{
		"TIP_STATE_DIR":    stateDir,
		"TIP_ADAPTER_PATH": adapterDir,
	}
	stdout, _, code := runTip(t,
		[]string{"observe",
			"--cmd-line", "docker build -t myapp .",
			"--exit", "1"},
		env,
	)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}
	if !strings.Contains(stdout, "fix:") {
		t.Errorf("missing fix: prefix, got: %q", stdout)
	}
	if !strings.Contains(stdout, "docker login") {
		t.Errorf("missing recovery text, got: %q", stdout)
	}
}

func TestObserveErrorNoAdapter(t *testing.T) {
	stateDir := freshStateDir(t)
	env := map[string]string{
		"TIP_STATE_DIR":    stateDir,
		"TIP_ADAPTER_PATH": t.TempDir(),
	}
	stdout, _, code := runTip(t,
		[]string{"observe",
			"--cmd-line", "badcmd --flag",
			"--exit", "1"},
		env,
	)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}
	// No adapter + error = no output
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("expected no output, got: %q", stdout)
	}
}
