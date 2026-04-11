package e2e_test

import (
	"strings"
	"testing"
)

func TestInstallBareName(t *testing.T) {
	stdout, _, code := runTip(t, []string{"install", "git"}, nil)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}
	if !strings.Contains(stdout, "hop.top/tip-git@latest") {
		t.Errorf("expected hop.top/tip-git@latest, got: %q", stdout)
	}
}

func TestInstallFullURI(t *testing.T) {
	stdout, _, code := runTip(t,
		[]string{"install", "github.com/x/tip-y"}, nil)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}
	if !strings.Contains(stdout, "github.com/x/tip-y") {
		t.Errorf("expected passthrough URI, got: %q", stdout)
	}
}

func TestInstallWithVersion(t *testing.T) {
	stdout, _, code := runTip(t,
		[]string{"install", "git@0.1.0"}, nil)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}
	if !strings.Contains(stdout, "hop.top/tip-git@v0.1.0") {
		t.Errorf("expected versioned URI, got: %q", stdout)
	}
}

func TestAdapterList(t *testing.T) {
	adapterDir := mockAdapterDir(t, map[string]map[string]any{
		"git":    {"suggestions": []any{}},
		"docker": {"suggestions": []any{}},
	})
	env := map[string]string{
		"TIP_ADAPTER_PATH": adapterDir,
	}

	stdout, _, code := runTip(t,
		[]string{"adapter", "list"}, env)
	if code != 0 {
		t.Fatalf("exit code: %d", code)
	}
	if !strings.Contains(stdout, "git") {
		t.Error("missing git adapter in list")
	}
	if !strings.Contains(stdout, "docker") {
		t.Error("missing docker adapter in list")
	}
}
