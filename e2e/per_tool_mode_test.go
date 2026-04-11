package e2e_test

import (
	"strings"
	"testing"
)

func TestOnPerToolWithMode(t *testing.T) {
	stateDir := freshStateDir(t)
	env := map[string]string{
		"TIP_STATE_DIR": stateDir,
	}

	stdout, _, code := runTip(t,
		[]string{"on", "git", "--mode", "guided"}, env)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}
	if !strings.Contains(stdout, "git") {
		t.Errorf("expected git in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "guided") {
		t.Errorf("expected guided in output, got: %q", stdout)
	}
}

func TestOffPerTool(t *testing.T) {
	stateDir := freshStateDir(t)
	env := map[string]string{
		"TIP_STATE_DIR": stateDir,
	}

	stdout, _, code := runTip(t,
		[]string{"off", "docker"}, env)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}
	if !strings.Contains(stdout, "docker") {
		t.Errorf("expected docker in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "disabled") {
		t.Errorf("expected disabled msg, got: %q", stdout)
	}
}

func TestStatusShowsToolOverrides(t *testing.T) {
	stateDir := freshStateDir(t)
	env := map[string]string{
		"TIP_STATE_DIR": stateDir,
	}

	// Set per-tool overrides
	runTip(t, []string{"on", "git", "--mode", "guided"}, env)
	runTip(t, []string{"off", "docker"}, env)

	stdout, _, code := runTip(t, []string{"status"}, env)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}
	if !strings.Contains(stdout, "tool overrides") {
		t.Error("expected tool overrides section")
	}
	if !strings.Contains(stdout, "git") {
		t.Error("expected git override in status")
	}
	if !strings.Contains(stdout, "docker") {
		t.Error("expected docker override in status")
	}
}
