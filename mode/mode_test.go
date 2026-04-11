package mode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultIsGuided(t *testing.T) {
	m, err := Parse("")
	require.NoError(t, err)
	assert.Equal(t, Guided, m)
}

func TestParseGuided(t *testing.T) {
	m, err := Parse("guided")
	require.NoError(t, err)
	assert.Equal(t, Guided, m)
}

func TestParseTransparent(t *testing.T) {
	m, err := Parse("transparent")
	require.NoError(t, err)
	assert.Equal(t, Transparent, m)
}

func TestParseQuiet(t *testing.T) {
	m, err := Parse("quiet")
	require.NoError(t, err)
	assert.Equal(t, Quiet, m)
}

func TestParseCaseInsensitive(t *testing.T) {
	m, err := Parse("GUIDED")
	require.NoError(t, err)
	assert.Equal(t, Guided, m)
}

func TestParseUnknown(t *testing.T) {
	_, err := Parse("invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown mode")
}

func TestString(t *testing.T) {
	assert.Equal(t, "guided", Guided.String())
	assert.Equal(t, "transparent", Transparent.String())
	assert.Equal(t, "quiet", Quiet.String())
}

func TestLoadConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	tipDir := filepath.Join(dir, "tip")
	require.NoError(t, os.MkdirAll(tipDir, 0o750))
	require.NoError(t, os.WriteFile(
		filepath.Join(tipDir, "config.yaml"),
		[]byte("mode: transparent\n"),
		0o644,
	))

	cfg, err := LoadConfig()
	require.NoError(t, err)
	m, err := cfg.Resolve()
	require.NoError(t, err)
	assert.Equal(t, Transparent, m)
}

func TestLoadConfigMissingFileDefaultsGuided(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	m, err := cfg.Resolve()
	require.NoError(t, err)
	assert.Equal(t, Guided, m)
}
