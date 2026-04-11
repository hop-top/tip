package toggle_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"hop.top/tip/toggle"
)

func tmpStateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", dir)
	return dir
}

func TestOnSetsGlobalEnabled(t *testing.T) {
	tmpStateDir(t)
	s := toggle.NewStore()

	require.NoError(t, s.On("", ""))
	assert.True(t, s.IsEnabled(""))
}

func TestOffClearsGlobalEnabled(t *testing.T) {
	tmpStateDir(t)
	s := toggle.NewStore()

	require.NoError(t, s.On("", ""))
	require.NoError(t, s.Off(""))
	assert.False(t, s.IsEnabled(""))
}

func TestPerToolOnOff(t *testing.T) {
	tmpStateDir(t)
	s := toggle.NewStore()

	require.NoError(t, s.On("git", ""))
	assert.True(t, s.IsEnabled("git"))
	assert.False(t, s.IsEnabled("docker"))

	require.NoError(t, s.Off("git"))
	assert.False(t, s.IsEnabled("git"))
}

func TestModeOverride(t *testing.T) {
	tmpStateDir(t)
	s := toggle.NewStore()

	require.NoError(t, s.On("git", "quiet"))
	m := s.Mode("git")
	assert.Equal(t, "quiet", m)
}

func TestStatePersists(t *testing.T) {
	dir := tmpStateDir(t)
	s := toggle.NewStore()
	require.NoError(t, s.On("", "guided"))

	// Verify file exists
	f := filepath.Join(dir, "tip", "state.json")
	_, err := os.Stat(f)
	assert.NoError(t, err)

	// Reload
	s2 := toggle.NewStore()
	assert.True(t, s2.IsEnabled(""))
	assert.Equal(t, "guided", s2.Mode(""))
}

func TestDefaultDisabled(t *testing.T) {
	tmpStateDir(t)
	s := toggle.NewStore()
	assert.False(t, s.IsEnabled(""))
}
