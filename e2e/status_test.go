package e2e_test

import (
	"strings"
	"testing"
)

func TestStatusOutput(t *testing.T) {
	stateDir := freshStateDir(t)
	adapterDir := mockAdapterDir(t, map[string]map[string]any{
		"git": {"suggestions": []any{}},
	})
	env := map[string]string{
		"TIP_STATE_DIR":    stateDir,
		"TIP_ADAPTER_PATH": adapterDir,
	}

	stdout, _, code := runTip(t, []string{"status"}, env)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}

	// Must include mode
	if !strings.Contains(stdout, "mode:") {
		t.Error("missing mode in status output")
	}
	// Must include adapters
	if !strings.Contains(stdout, "adapters:") {
		t.Error("missing adapters in status output")
	}
	// Must include session info
	if !strings.Contains(stdout, "session:") {
		t.Error("missing session in status output")
	}
	// Should show git adapter
	if !strings.Contains(stdout, "git") {
		t.Error("missing git adapter in status")
	}
}

func TestStatusDefaultMode(t *testing.T) {
	stateDir := freshStateDir(t)
	env := map[string]string{
		"TIP_STATE_DIR": stateDir,
	}

	stdout, _, _ := runTip(t, []string{"status"}, env)
	if !strings.Contains(stdout, "mode: guided") {
		t.Errorf("expected default guided mode, got: %q", stdout)
	}
}

func TestStatusAfterDisable(t *testing.T) {
	stateDir := freshStateDir(t)
	env := map[string]string{
		"TIP_STATE_DIR": stateDir,
	}

	runTip(t, []string{"off"}, env)
	stdout, _, _ := runTip(t, []string{"status"}, env)
	if !strings.Contains(stdout, "enabled: off") {
		t.Errorf("expected enabled: off, got: %q", stdout)
	}
}
