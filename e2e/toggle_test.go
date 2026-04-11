package e2e_test

import (
	"strings"
	"testing"
)

func TestToggleOff(t *testing.T) {
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
	}

	// Disable
	out, _, code := runTip(t, []string{"off"}, env)
	if code != 0 {
		t.Fatalf("tip off exit: %d", code)
	}
	if !strings.Contains(out, "disabled") {
		t.Errorf("expected disabled msg, got: %q", out)
	}

	// Observe should produce nothing
	stdout, _, code := runTip(t,
		[]string{"observe", "--cmd-line", "git commit -m fix",
			"--exit", "0"},
		env,
	)
	if code != 0 {
		t.Fatalf("observe exit: %d", code)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("expected no output when off, got: %q", stdout)
	}
}

func TestToggleOnReenables(t *testing.T) {
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
	}

	// Off then on
	runTip(t, []string{"off"}, env)
	out, _, code := runTip(t, []string{"on"}, env)
	if code != 0 {
		t.Fatalf("tip on exit: %d", code)
	}
	if !strings.Contains(out, "enabled") {
		t.Errorf("expected enabled msg, got: %q", out)
	}

	// Observe should produce output again
	stdout, _, code := runTip(t,
		[]string{"observe", "--cmd-line", "git commit -m fix",
			"--exit", "0"},
		env,
	)
	if code != 0 {
		t.Fatalf("observe exit: %d", code)
	}
	if !strings.Contains(stdout, "git push") {
		t.Errorf("expected suggestion after re-enable, got: %q",
			stdout)
	}
}

func TestToggleSessionScoped(t *testing.T) {
	stateDir := freshStateDir(t)
	env := map[string]string{
		"TIP_STATE_DIR": stateDir,
	}

	// Disable via state file
	runTip(t, []string{"off"}, env)

	// Different TIP_STATE_DIR = different session scope
	stateDir2 := freshStateDir(t)
	env2 := map[string]string{
		"TIP_STATE_DIR":    stateDir2,
		"TIP_ADAPTER_PATH": t.TempDir(),
	}

	// Should be enabled by default in session 2
	out, _, _ := runTip(t, []string{"status"}, env2)
	if strings.Contains(out, "enabled: off") {
		t.Error("second session should not be affected by first")
	}
}
