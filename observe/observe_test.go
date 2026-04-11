package observe

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tipctx "hop.top/tip/context"
	"hop.top/tip/mode"
	"hop.top/tip/policy"
	"hop.top/tip/session"
)

// mockAdapter returns canned suggestions.
type mockAdapter struct {
	suggestions []policy.Suggestion
	isGeneric   bool
	err         error
	called      bool
}

func (m *mockAdapter) Suggest(env tipctx.Envelope) ([]policy.Suggestion, bool, error) {
	m.called = true
	return m.suggestions, m.isGeneric, m.err
}

// mockRenderer records whether Render was called.
type mockRenderer struct {
	called bool
	result policy.Result
}

func (m *mockRenderer) Render(r policy.Result) error {
	m.called = true
	m.result = r
	return nil
}

func newTestStore(t *testing.T) *session.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := session.OpenAt(path)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

func TestRunWithMockAdapter(t *testing.T) {
	store := newTestStore(t)
	adapter := &mockAdapter{
		suggestions: []policy.Suggestion{
			{Text: "git push", Confidence: 0.9, Primary: true},
			{Text: "git log", Confidence: 0.4},
		},
	}
	renderer := &mockRenderer{}

	err := Run(Params{
		CmdLine:  "git commit -m 'wip'",
		ExitCode: 0,
		Cwd:      "/tmp/test",
		Mode:     mode.Guided,
		Store:    store,
		Adapter:  adapter,
		Renderer: renderer,
	})
	require.NoError(t, err)
	assert.True(t, adapter.called)
	assert.True(t, renderer.called)
	require.NotNil(t, renderer.result.Primary)
	assert.Equal(t, "git push", renderer.result.Primary.Text)
}

func TestRunNoAdapter(t *testing.T) {
	store := newTestStore(t)
	renderer := &mockRenderer{}

	err := Run(Params{
		CmdLine:  "git status",
		ExitCode: 0,
		Cwd:      "/tmp/test",
		Mode:     mode.Guided,
		Store:    store,
		Adapter:  nil,
		Renderer: renderer,
	})
	require.NoError(t, err)
	// No adapter = no suggestions = renderer still called but with
	// empty result.
	assert.True(t, renderer.called)
	assert.Nil(t, renderer.result.Primary)
}

func TestRunTransparentSkipsRender(t *testing.T) {
	store := newTestStore(t)
	adapter := &mockAdapter{
		suggestions: []policy.Suggestion{
			{Text: "git push", Confidence: 0.9, Primary: true},
		},
	}
	renderer := &mockRenderer{}

	err := Run(Params{
		CmdLine:  "git commit -m 'wip'",
		ExitCode: 0,
		Cwd:      "/tmp/test",
		Mode:     mode.Transparent,
		Store:    store,
		Adapter:  adapter,
		Renderer: renderer,
	})
	require.NoError(t, err)
	assert.True(t, adapter.called)
	assert.False(t, renderer.called, "transparent should skip render")
}

func TestRunQuietSkipsEverything(t *testing.T) {
	adapter := &mockAdapter{
		suggestions: []policy.Suggestion{
			{Text: "git push", Confidence: 0.9},
		},
	}
	renderer := &mockRenderer{}

	err := Run(Params{
		CmdLine:  "git commit",
		ExitCode: 0,
		Cwd:      "/tmp/test",
		Mode:     mode.Quiet,
		Adapter:  adapter,
		Renderer: renderer,
	})
	require.NoError(t, err)
	assert.False(t, adapter.called, "quiet should not call adapter")
	assert.False(t, renderer.called, "quiet should not render")
}

func TestRunAdapterError(t *testing.T) {
	store := newTestStore(t)
	adapter := &mockAdapter{
		err: errors.New("adapter timeout"),
	}
	renderer := &mockRenderer{}

	err := Run(Params{
		CmdLine:  "docker build .",
		ExitCode: 0,
		Cwd:      "/tmp/test",
		Mode:     mode.Guided,
		Store:    store,
		Adapter:  adapter,
		Renderer: renderer,
	})
	// Adapter errors are non-fatal.
	require.NoError(t, err)
	assert.True(t, renderer.called)
}

func TestRunEmptyCommand(t *testing.T) {
	renderer := &mockRenderer{}
	err := Run(Params{
		CmdLine:  "",
		ExitCode: 0,
		Mode:     mode.Guided,
		Renderer: renderer,
	})
	require.NoError(t, err)
	assert.False(t, renderer.called, "empty cmd should bail early")
}

func TestRunRecordsToSession(t *testing.T) {
	store := newTestStore(t)
	adapter := &mockAdapter{}

	err := Run(Params{
		CmdLine:  "git status",
		ExitCode: 0,
		Cwd:      "/tmp/proj",
		Mode:     mode.Guided,
		Store:    store,
		Adapter:  adapter,
	})
	require.NoError(t, err)

	// Verify command was recorded.
	sess, err := store.FindOrCreateSession(ctxBackground(), "/tmp/proj")
	require.NoError(t, err)
	history, err := store.History(ctxBackground(), sess.ID, 10)
	require.NoError(t, err)
	require.Len(t, history, 1)
	assert.Equal(t, "git", history[0].Cmd)
	assert.Equal(t, []string{"status"}, history[0].Args)
}
