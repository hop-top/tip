package status_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"hop.top/tip/status"
	"hop.top/tip/toggle"
)

func TestShowsMode(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	s := toggle.NewStore()
	_ = s.On("", "guided")

	var buf bytes.Buffer
	status.Write(&buf, status.Info{
		Store:    s,
		Adapters: nil,
	})
	out := buf.String()
	assert.Contains(t, out, "guided")
}

func TestShowsPerToolOverrides(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	s := toggle.NewStore()
	_ = s.On("git", "quiet")

	var buf bytes.Buffer
	status.Write(&buf, status.Info{Store: s})
	out := buf.String()
	assert.Contains(t, out, "git")
	assert.Contains(t, out, "quiet")
}

func TestShowsAdapterList(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	s := toggle.NewStore()

	var buf bytes.Buffer
	status.Write(&buf, status.Info{
		Store:    s,
		Adapters: []string{"tip-git", "tip-docker"},
	})
	out := buf.String()
	assert.Contains(t, out, "tip-git")
	assert.Contains(t, out, "tip-docker")
}

func TestShowsSessionCount(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	s := toggle.NewStore()

	var buf bytes.Buffer
	status.Write(&buf, status.Info{
		Store:        s,
		CommandCount: 42,
	})
	out := buf.String()
	assert.Contains(t, out, "42")
}

func TestDisabledState(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	s := toggle.NewStore()

	var buf bytes.Buffer
	status.Write(&buf, status.Info{Store: s})
	out := buf.String()
	assert.Contains(t, out, "disabled")
}
