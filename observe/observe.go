// Package observe orchestrates the tip observation pipeline.
//
// Flow: parse -> session -> adapter -> policy -> render.
// Adapter and Renderer are interfaces; concrete implementations
// live in separate packages/tracks.
package observe

import (
	"fmt"
	"log"

	tipctx "hop.top/tip/context"
	"hop.top/tip/mode"
	"hop.top/tip/policy"
	"hop.top/tip/session"
)

// Adapter queries an external or built-in adapter for suggestions.
type Adapter interface {
	// Suggest returns suggestions for the given context envelope.
	// isGeneric should be true if this is the built-in generic adapter.
	Suggest(env tipctx.Envelope) ([]policy.Suggestion, bool, error)
}

// Renderer displays suggestions to the user.
type Renderer interface {
	// Render displays a policy result to the terminal.
	Render(result policy.Result) error
}

// Params holds the inputs to the observe pipeline.
type Params struct {
	CmdLine  string
	ExitCode int
	Stderr   string
	Cwd      string
	Mode     mode.Mode
	Store    *session.Store
	Adapter  Adapter  // nil = no adapter available.
	Renderer Renderer // nil = skip render.
}

// Run executes the full observation pipeline.
func Run(p Params) error {
	// Quiet mode: skip everything.
	if p.Mode == mode.Quiet {
		return nil
	}

	// Parse command line.
	parsed := tipctx.Parse(p.CmdLine)
	if parsed.Cmd == "" {
		return nil
	}

	// Session: find-or-create, load history.
	var history []session.Command
	var sess *session.Session
	if p.Store != nil {
		var err error
		sess, err = p.Store.FindOrCreateSession(
			ctxBackground(), p.Cwd,
		)
		if err != nil {
			return fmt.Errorf("session: %w", err)
		}
		// Best-effort: history load failure is non-fatal since observe
		// can still produce useful output without prior commands.
		history, err = p.Store.History(ctxBackground(), sess.ID, 10)
		if err != nil {
			log.Printf("tip: load history (best-effort): %v", err)
		}
	}

	// Build context envelope.
	env := tipctx.Build(parsed, p.ExitCode, p.Stderr, p.Cwd, history)

	// Query adapter.
	var suggestions []policy.Suggestion
	var isGeneric bool
	if p.Adapter != nil {
		var err error
		suggestions, isGeneric, err = p.Adapter.Suggest(env)
		if err != nil {
			// Adapter errors are non-fatal; continue without suggestions.
			suggestions = nil
		}
	}

	// Apply policy.
	result := policy.Apply(suggestions, policy.Options{
		IsGeneric:   isGeneric,
		OriginalCmd: p.CmdLine,
	})

	// Record command in session.
	// Best-effort: command recording failure is non-fatal; the primary
	// value of observe is the suggestion output, not the audit trail.
	if p.Store != nil && sess != nil {
		if err := p.Store.AppendCommand(ctxBackground(), sess.ID, session.Command{
			Cmd:      parsed.Cmd,
			Args:     parsed.Args,
			ExitCode: p.ExitCode,
		}); err != nil {
			log.Printf("tip: append command (best-effort): %v", err)
		}
	}

	// Transparent mode: record but skip render.
	if p.Mode == mode.Transparent {
		return nil
	}

	// Render (guided mode).
	if p.Renderer != nil {
		return p.Renderer.Render(result)
	}
	return nil
}
