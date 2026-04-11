package session

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := OpenAt(path)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

func TestOpenCreatesTables(t *testing.T) {
	s := newTestStore(t)
	db := s.DB()

	// Verify all three tables exist.
	for _, table := range []string{"sessions", "commands", "adapter_cache"} {
		var name string
		err := db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`,
			table,
		).Scan(&name)
		require.NoError(t, err, "table %s should exist", table)
		assert.Equal(t, table, name)
	}
}

func TestFindOrCreateSession(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Create new session.
	sess1, err := s.FindOrCreateSession(ctx, "/tmp/project")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/project", sess1.Cwd)
	assert.NotEmpty(t, sess1.ID)

	// Find existing session.
	sess2, err := s.FindOrCreateSession(ctx, "/tmp/project")
	require.NoError(t, err)
	assert.Equal(t, sess1.ID, sess2.ID)
}

func TestFindOrCreateSessionDifferentCwd(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	sess1, err := s.FindOrCreateSession(ctx, "/tmp/a")
	require.NoError(t, err)
	sess2, err := s.FindOrCreateSession(ctx, "/tmp/b")
	require.NoError(t, err)
	assert.NotEqual(t, sess1.ID, sess2.ID)
}

func TestAppendCommandAndHistory(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	sess, err := s.FindOrCreateSession(ctx, "/tmp/work")
	require.NoError(t, err)

	// Append commands.
	for _, c := range []Command{
		{Cmd: "git", Args: []string{"add", "."}, ExitCode: 0},
		{Cmd: "git", Args: []string{"commit", "-m", "wip"}, ExitCode: 0},
		{Cmd: "git", Args: []string{"push"}, ExitCode: 1},
	} {
		err := s.AppendCommand(ctx, sess.ID, c)
		require.NoError(t, err)
	}

	// Query last 10 (should get all 3).
	history, err := s.History(ctx, sess.ID, 10)
	require.NoError(t, err)
	require.Len(t, history, 3)

	// Oldest first.
	assert.Equal(t, "git", history[0].Cmd)
	assert.Equal(t, []string{"add", "."}, history[0].Args)
	assert.Equal(t, []string{"push"}, history[2].Args)
	assert.Equal(t, 1, history[2].ExitCode)
}

func TestHistoryLimit(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	sess, err := s.FindOrCreateSession(ctx, "/tmp/limit")
	require.NoError(t, err)

	// Insert 15 commands.
	for i := 0; i < 15; i++ {
		err := s.AppendCommand(ctx, sess.ID, Command{
			Cmd: "echo", Args: []string{"hello"}, ExitCode: 0,
		})
		require.NoError(t, err)
	}

	// Query last 10.
	history, err := s.History(ctx, sess.ID, 10)
	require.NoError(t, err)
	assert.Len(t, history, 10)
}
