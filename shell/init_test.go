package shell_test

import (
	"strings"
	"testing"

	"hop.top/tip/shell"
)

func TestZshHookOutput(t *testing.T) {
	out := shell.InitCode("zsh")

	required := []string{
		"__tip_preexec",
		"__tip_precmd",
		"add-zsh-hook",
		"__tip_enabled",
		"tip observe",
	}
	for _, s := range required {
		if !strings.Contains(out, s) {
			t.Errorf("zsh init missing %q", s)
		}
	}
}

func TestBashHookOutput(t *testing.T) {
	out := shell.InitCode("bash")

	required := []string{
		"DEBUG",
		"PROMPT_COMMAND",
		"tip observe",
		"__tip_preexec",
		"__tip_precmd",
	}
	for _, s := range required {
		if !strings.Contains(out, s) {
			t.Errorf("bash init missing %q", s)
		}
	}
}

func TestUnsupportedShell(t *testing.T) {
	out := shell.InitCode("fish")
	if out != "" {
		t.Errorf("expected empty output for unsupported shell, got %q", out)
	}
}

func TestZshSyntaxBasic(t *testing.T) {
	out := shell.InitCode("zsh")
	// Balanced braces
	if strings.Count(out, "{") != strings.Count(out, "}") {
		t.Error("unbalanced braces in zsh output")
	}
	// Every function definition has a closing brace
	if strings.Count(out, "() {") < 2 {
		t.Error("expected at least 2 function definitions in zsh output")
	}
}

func TestBashSyntaxBasic(t *testing.T) {
	out := shell.InitCode("bash")
	if strings.Count(out, "{") != strings.Count(out, "}") {
		t.Error("unbalanced braces in bash output")
	}
	if strings.Count(out, "() {") < 2 {
		t.Error("expected at least 2 function definitions in bash output")
	}
}

func TestZshCapturesExitCode(t *testing.T) {
	out := shell.InitCode("zsh")
	if !strings.Contains(out, "$?") {
		t.Error("zsh hook should capture exit code via $?")
	}
}

func TestBashCapturesExitCode(t *testing.T) {
	out := shell.InitCode("bash")
	if !strings.Contains(out, "$?") {
		t.Error("bash hook should capture exit code via $?")
	}
}
