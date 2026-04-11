// Package session provides SQLite-backed session persistence for tip.
//
// A session groups commands executed in the same working directory.
// Sessions are found-or-created by cwd, and commands are appended
// with metadata (intent, stage, adapter) for history queries.
package session

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"hop.top/kit/sqlstore"
	"hop.top/kit/xdg"
)

// migrateSQL creates the sessions, commands, and adapter_cache tables.
const migrateSQL = `
CREATE TABLE IF NOT EXISTS sessions (
  id          TEXT PRIMARY KEY,
  cwd         TEXT NOT NULL,
  created_at  INTEGER NOT NULL,
  updated_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS commands (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id  TEXT NOT NULL REFERENCES sessions(id),
  cmd         TEXT NOT NULL,
  args        TEXT NOT NULL,
  exit_code   INTEGER NOT NULL,
  intent      TEXT,
  stage       TEXT,
  adapter     TEXT,
  created_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS adapter_cache (
  cmd         TEXT PRIMARY KEY,
  describe    TEXT NOT NULL,
  help_tree   TEXT,
  cached_at   INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_commands_session
  ON commands(session_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_cwd ON sessions(cwd);
`

// Command represents a single command execution stored in a session.
type Command struct {
	ID        int64    `json:"id"`
	SessionID string   `json:"session_id"`
	Cmd       string   `json:"cmd"`
	Args      []string `json:"args"`
	ExitCode  int      `json:"exit_code"`
	Intent    string   `json:"intent,omitempty"`
	Stage     string   `json:"stage,omitempty"`
	Adapter   string   `json:"adapter,omitempty"`
	CreatedAt int64    `json:"created_at"`
}

// Session represents a tip session scoped to a working directory.
type Session struct {
	ID        string `json:"id"`
	Cwd       string `json:"cwd"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// Store wraps kit/sqlstore for session persistence.
type Store struct {
	kv *sqlstore.Store
}

// Open creates or opens the session database at the XDG data path.
func Open() (*Store, error) {
	dataDir, err := xdg.DataDir("tip")
	if err != nil {
		return nil, fmt.Errorf("resolve data dir: %w", err)
	}
	return OpenAt(dataDir + "/sessions.db")
}

// OpenAt creates or opens the session database at the given path.
func OpenAt(path string) (*Store, error) {
	kv, err := sqlstore.Open(path, sqlstore.Options{
		MigrateSQL: migrateSQL,
	})
	if err != nil {
		return nil, fmt.Errorf("open session store: %w", err)
	}
	return &Store{kv: kv}, nil
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.kv.Close() }

// DB returns the underlying *sql.DB for direct queries.
func (s *Store) DB() *sql.DB { return s.kv.DB() }

// FindOrCreateSession returns the session for the given cwd, creating
// one if it does not exist.
func (s *Store) FindOrCreateSession(
	ctx context.Context, cwd string,
) (*Session, error) {
	db := s.kv.DB()
	now := time.Now().Unix()

	var sess Session
	err := db.QueryRowContext(ctx,
		`SELECT id, cwd, created_at, updated_at FROM sessions WHERE cwd = ?`,
		cwd,
	).Scan(&sess.ID, &sess.Cwd, &sess.CreatedAt, &sess.UpdatedAt)

	if err == sql.ErrNoRows {
		id := newID()
		_, err = db.ExecContext(ctx,
			`INSERT INTO sessions (id, cwd, created_at, updated_at)
			 VALUES (?, ?, ?, ?)`,
			id, cwd, now, now,
		)
		if err != nil {
			return nil, fmt.Errorf("create session: %w", err)
		}
		return &Session{ID: id, Cwd: cwd, CreatedAt: now, UpdatedAt: now}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}

	// Touch updated_at.
	_, _ = db.ExecContext(ctx,
		`UPDATE sessions SET updated_at = ? WHERE id = ?`, now, sess.ID)
	sess.UpdatedAt = now
	return &sess, nil
}

// AppendCommand stores a command execution in the given session.
func (s *Store) AppendCommand(
	ctx context.Context, sessionID string, cmd Command,
) error {
	argsJSON, err := json.Marshal(cmd.Args)
	if err != nil {
		return fmt.Errorf("marshal args: %w", err)
	}
	now := time.Now().Unix()
	_, err = s.kv.DB().ExecContext(ctx,
		`INSERT INTO commands
		 (session_id, cmd, args, exit_code, intent, stage, adapter, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sessionID, cmd.Cmd, string(argsJSON), cmd.ExitCode,
		cmd.Intent, cmd.Stage, cmd.Adapter, now,
	)
	return err
}

// History returns the last n commands for the given session, ordered
// oldest-first.
func (s *Store) History(
	ctx context.Context, sessionID string, n int,
) ([]Command, error) {
	rows, err := s.kv.DB().QueryContext(ctx,
		`SELECT id, session_id, cmd, args, exit_code,
		        intent, stage, adapter, created_at
		 FROM commands
		 WHERE session_id = ?
		 ORDER BY created_at DESC, id DESC
		 LIMIT ?`,
		sessionID, n,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []Command
	for rows.Next() {
		var c Command
		var argsJSON string
		var intent, stage, adapter sql.NullString
		if err := rows.Scan(
			&c.ID, &c.SessionID, &c.Cmd, &argsJSON,
			&c.ExitCode, &intent, &stage, &adapter, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(argsJSON), &c.Args)
		c.Intent = intent.String
		c.Stage = stage.String
		c.Adapter = adapter.String
		cmds = append(cmds, c)
	}
	// Reverse to oldest-first.
	for i, j := 0, len(cmds)-1; i < j; i, j = i+1, j-1 {
		cmds[i], cmds[j] = cmds[j], cmds[i]
	}
	return cmds, rows.Err()
}

// newID returns a short random hex ID prefixed with "s-".
func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "s-" + hex.EncodeToString(b)
}
