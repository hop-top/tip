package e2e_test

import (
	"strings"
	"testing"
)

func TestInitZsh(t *testing.T) {
	stdout, _, code := runTip(t, []string{"init", "zsh"}, nil)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}

	required := []string{
		"__tip_preexec",
		"__tip_precmd",
		"add-zsh-hook",
		"autoload -Uz add-zsh-hook",
		"add-zsh-hook preexec __tip_preexec",
		"add-zsh-hook precmd __tip_precmd",
	}
	for _, s := range required {
		if !strings.Contains(stdout, s) {
			t.Errorf("zsh hook missing %q", s)
		}
	}
}

func TestInitBash(t *testing.T) {
	stdout, _, code := runTip(t, []string{"init", "bash"}, nil)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}

	required := []string{
		"PROMPT_COMMAND",
		"__tip_precmd",
		"__tip_preexec",
		"trap",
	}
	for _, s := range required {
		if !strings.Contains(stdout, s) {
			t.Errorf("bash hook missing %q", s)
		}
	}
}

func TestInitUnsupportedShell(t *testing.T) {
	_, stderr, code := runTip(t, []string{"init", "fish"}, nil)
	if code == 0 {
		t.Fatal("expected non-zero exit for unsupported shell")
	}
	if !strings.Contains(stderr, "unsupported") {
		t.Errorf("expected unsupported error, got: %s", stderr)
	}
}
