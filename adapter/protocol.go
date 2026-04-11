// Package adapter provides adapter discovery, invocation, and protocol types
// for the tip CLI observation layer.
package adapter

// DescribeResponse is returned by `tip-<tool> describe` on stdout.
type DescribeResponse struct {
	Name           string   `json:"name"`
	Version        string   `json:"version"`
	ToolVersionCmd string   `json:"tool_version_cmd,omitempty"`
	Commands       []string `json:"commands,omitempty"`
	Capabilities   []string `json:"capabilities,omitempty"`
}

// SuggestResponse is returned by `tip-<tool> suggest` on stdout.
type SuggestResponse struct {
	Intent      string       `json:"intent"`
	Stage       string       `json:"stage"`
	Suggestions []Suggestion `json:"suggestions"`
	Recovery    *Suggestion  `json:"recovery,omitempty"`
}

// Suggestion represents a single next-step suggestion.
type Suggestion struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Primary    bool    `json:"primary,omitempty"`
	Reason     string  `json:"reason,omitempty"`
}

// ContextEnvelope is piped to `tip-<tool> suggest` on stdin.
type ContextEnvelope struct {
	Cmd      string         `json:"cmd"`
	Args     []string       `json:"args"`
	ExitCode int            `json:"exit_code"`
	Stderr   string         `json:"stderr,omitempty"`
	Cwd      string         `json:"cwd,omitempty"`
	History  []HistoryEntry `json:"history,omitempty"`
}

// HistoryEntry is a single command in the session history.
type HistoryEntry struct {
	Cmd      string   `json:"cmd"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}
