package version_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"hop.top/tip/version"
)

func TestVersionOutputSemver(t *testing.T) {
	var buf bytes.Buffer
	version.Write(&buf, version.Info{
		Version: "0.0.1-alpha.1",
		Commit:  "abc1234",
		Date:    "2025-04-10",
	})
	out := buf.String()
	assert.Contains(t, out, "0.0.1-alpha.1")
	assert.Contains(t, out, "abc1234")
	assert.Contains(t, out, "2025-04-10")
}

func TestVersionDevDefaults(t *testing.T) {
	var buf bytes.Buffer
	version.Write(&buf, version.Info{
		Version: "dev",
		Commit:  "none",
		Date:    "unknown",
	})
	out := buf.String()
	assert.Contains(t, out, "dev")
}

func TestVersionCmdExists(t *testing.T) {
	cmd := version.Cmd(version.Info{
		Version: "1.0.0",
		Commit:  "abc",
		Date:    "today",
	})
	assert.Equal(t, "version", cmd.Use)
}

func TestUpgradeCmdExists(t *testing.T) {
	cmd := version.UpgradeCmd("tip", "1.0.0")
	assert.Equal(t, "upgrade", cmd.Use)
}
