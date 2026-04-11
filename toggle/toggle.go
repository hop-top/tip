// Package toggle manages on/off state for tip, globally and per-tool.
//
// State is persisted as JSON in XDG_RUNTIME_DIR/tip/state.json.
// The shell hook checks this state before calling tip observe.
package toggle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// ToolState holds the enabled/mode state for a single tool (or global).
type ToolState struct {
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode,omitempty"`
}

// state is the full persisted state.
type state struct {
	Global ToolState            `json:"global"`
	Tools  map[string]ToolState `json:"tools,omitempty"`
}

// Store manages tip on/off state.
type Store struct {
	mu   sync.Mutex
	path string
	data state
}

// NewStore creates a Store backed by the XDG runtime dir.
func NewStore() *Store {
	s := &Store{
		path: statePath(),
		data: state{Tools: make(map[string]ToolState)},
	}
	s.load()
	return s
}

// On enables tip globally (tool=="") or for a specific tool.
// mode is optional; empty means default mode.
func (s *Store) On(tool, mode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ts := ToolState{Enabled: true, Mode: mode}
	if tool == "" {
		s.data.Global = ts
	} else {
		s.data.Tools[tool] = ts
	}
	return s.save()
}

// Off disables tip globally (tool=="") or for a specific tool.
func (s *Store) Off(tool string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tool == "" {
		s.data.Global = ToolState{}
	} else {
		s.data.Tools[tool] = ToolState{Enabled: false}
	}
	return s.save()
}

// IsEnabled returns whether tip is enabled for the given tool.
// Empty tool checks global state.
func (s *Store) IsEnabled(tool string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tool == "" {
		return s.data.Global.Enabled
	}
	ts, ok := s.data.Tools[tool]
	if ok {
		return ts.Enabled
	}
	return s.data.Global.Enabled
}

// Mode returns the mode override for the given tool (or global).
func (s *Store) Mode(tool string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tool == "" {
		return s.data.Global.Mode
	}
	ts, ok := s.data.Tools[tool]
	if ok {
		return ts.Mode
	}
	return ""
}

// State returns a snapshot of all tool states for display.
func (s *Store) State() (global ToolState, tools map[string]ToolState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := make(map[string]ToolState, len(s.data.Tools))
	for k, v := range s.data.Tools {
		cp[k] = v
	}
	return s.data.Global, cp
}

func statePath() string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "tip", "state.json")
}

func (s *Store) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	var st state
	if json.Unmarshal(data, &st) == nil {
		if st.Tools == nil {
			st.Tools = make(map[string]ToolState)
		}
		s.data = st
	}
}

func (s *Store) save() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(s.data)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}
