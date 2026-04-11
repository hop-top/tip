// Package context builds the context envelope piped to adapters.
//
// The envelope carries the parsed command, exit code, stderr, cwd,
// and recent session history in a JSON-serialisable struct.
package context

import (
	"hop.top/tip/session"
)

// HistoryEntry is a single command in the envelope's history array.
type HistoryEntry struct {
	Cmd      string   `json:"cmd"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

// Envelope is the JSON payload sent to adapter suggest stdin.
type Envelope struct {
	Cmd      string         `json:"cmd"`
	Args     []string       `json:"args"`
	ExitCode int            `json:"exit_code"`
	Stderr   string         `json:"stderr"`
	Cwd      string         `json:"cwd"`
	History  []HistoryEntry `json:"history"`
}

// Build constructs an Envelope from a parsed command, exit code, stderr,
// cwd, and session history.
func Build(parsed Parsed, exitCode int, stderr, cwd string, history []session.Command) Envelope {
	entries := make([]HistoryEntry, 0, len(history))
	for _, h := range history {
		entries = append(entries, HistoryEntry{
			Cmd:      h.Cmd,
			Args:     h.Args,
			ExitCode: h.ExitCode,
		})
	}
	return Envelope{
		Cmd:      parsed.Cmd,
		Args:     parsed.Args,
		ExitCode: exitCode,
		Stderr:   stderr,
		Cwd:      cwd,
		History:  entries,
	}
}
