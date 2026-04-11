package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionFlag(t *testing.T) {
	// Build the binary.
	bin := t.TempDir() + "/tip"
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = "."
	out, err := build.CombinedOutput()
	require.NoError(t, err, "build failed: %s", out)

	// Run with --version.
	cmd := exec.Command(bin, "--version")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "version flag failed: %s", output)
	require.Contains(t, string(output), "dev")
}
