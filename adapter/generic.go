package adapter

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"time"

	kitllm "hop.top/kit/llm"
	"hop.top/kit/toolspec"
	helpSource "hop.top/kit/toolspec/sources/help"
	llmSource "hop.top/kit/toolspec/sources/llm"
)

// MaxGenericConfidence caps confidence for generic adapter suggestions.
const MaxGenericConfidence = 0.6

// Generic is the built-in Level 0 adapter that works with any CLI tool.
// It uses a toolspec.Registry to resolve tool specs (subcommands, workflows,
// error patterns) and returns suggestions with confidence capped at
// MaxGenericConfidence.
type Generic struct {
	mu        sync.RWMutex
	cache     map[string][]string // legacy cache for Warm/CacheHit compat
	registry  *toolspec.Registry
	llmClient kitllm.Completer // nil = no LLM
	llmModel  string
}

// NewGeneric creates a new Generic adapter backed by a toolspec.Registry.
// If no registry is provided, a default one is created using the help source.
func NewGeneric(opts ...GenericOption) *Generic {
	g := &Generic{cache: make(map[string][]string)}
	for _, o := range opts {
		o(g)
	}
	if g.registry == nil {
		sources := []toolspec.RegistryOption{
			toolspec.WithSource(BuiltinWorkflows()),
			toolspec.WithSource(&helpSource.HelpSource{}),
		}
		// LLM source last — fills gaps left by static sources.
		// Results wrapped with enrichingSource to derive After
		// maps from Steps.
		if g.llmClient != nil {
			sources = append(sources, toolspec.WithSource(
				&enrichingSource{inner: llmSource.NewLLMSource(
					llmSource.Config{
						Client:  g.llmClient,
						Model:   g.llmModel,
						Enabled: true,
					},
				)},
			))
		}
		g.registry = toolspec.NewRegistry(sources...)
	}
	return g
}

// GenericOption configures a Generic adapter.
type GenericOption func(*Generic)

// WithRegistry sets a custom toolspec.Registry.
func WithRegistry(r *toolspec.Registry) GenericOption {
	return func(g *Generic) { g.registry = r }
}

// WithLLM enables the LLM fallback source for enriching workflows
// and error patterns. Results are cached by the registry.
func WithLLM(client kitllm.Completer, model string) GenericOption {
	return func(g *Generic) { g.llmClient = client; g.llmModel = model }
}

// Suggest returns suggestions based on the resolved ToolSpec for the command.
// It uses subcommands, workflows, and error patterns from the spec.
func (g *Generic) Suggest(envelope ContextEnvelope) *SuggestResponse {
	spec, err := g.registry.Resolve(envelope.Cmd)
	if err != nil || spec == nil {
		// Fall back to legacy cache if registry fails.
		return g.suggestFromCache(envelope)
	}

	var suggestions []Suggestion

	// 1. Workflow-aware: if the command matches a workflow step, suggest next steps.
	suggestions = append(suggestions, g.workflowSuggestions(spec, envelope)...)

	// 2. Error recovery: if exit code != 0 and stderr present, check error patterns.
	if envelope.ExitCode != 0 && envelope.Stderr != "" {
		if rec := g.errorRecovery(spec, envelope); rec != nil {
			return &SuggestResponse{
				Intent:      envelope.Cmd,
				Stage:       "generic",
				Suggestions: suggestions,
				Recovery:    rec,
			}
		}
	}

	// 3. Subcommand suggestions from spec.Commands.
	if len(suggestions) == 0 {
		suggestions = g.subcommandSuggestions(spec, envelope)
	}

	// If still nothing, try legacy cache (for Warm'd entries).
	if len(suggestions) == 0 {
		return g.suggestFromCache(envelope)
	}

	return &SuggestResponse{
		Intent:      envelope.Cmd,
		Stage:       "generic",
		Suggestions: suggestions,
	}
}

// subcommandSuggestions builds suggestions from spec.Commands.
func (g *Generic) subcommandSuggestions(spec *toolspec.ToolSpec, envelope ContextEnvelope) []Suggestion {
	if len(spec.Commands) == 0 {
		return nil
	}

	limit := min(3, len(spec.Commands))
	suggestions := make([]Suggestion, 0, limit)
	for i := 0; i < limit; i++ {
		conf := MaxGenericConfidence - float64(i)*0.1
		if conf < 0.1 {
			conf = 0.1
		}
		suggestions = append(suggestions, Suggestion{
			Text:       envelope.Cmd + " " + spec.Commands[i].Name,
			Confidence: conf,
			Primary:    i == 0,
		})
	}
	return suggestions
}

// workflowSuggestions checks if envelope.Cmd matches a workflow step and
// returns next-step suggestions from Workflow.After.
func (g *Generic) workflowSuggestions(spec *toolspec.ToolSpec, envelope ContextEnvelope) []Suggestion {
	// Build a full command string (cmd + first arg) for matching.
	fullCmd := envelope.Cmd
	if len(envelope.Args) > 0 {
		fullCmd = envelope.Cmd + " " + envelope.Args[0]
	}

	var suggestions []Suggestion
	for _, wf := range spec.Workflows {
		nextSteps, ok := wf.After[fullCmd]
		if !ok {
			// Try bare cmd without args.
			nextSteps, ok = wf.After[envelope.Cmd]
		}
		if !ok {
			continue
		}
		for i, step := range nextSteps {
			if i >= 3 {
				break
			}
			conf := MaxGenericConfidence - float64(i)*0.1
			if conf < 0.1 {
				conf = 0.1
			}
			suggestions = append(suggestions, Suggestion{
				Text:       step,
				Confidence: conf,
				Primary:    i == 0,
				Reason:     "workflow: " + wf.Name,
			})
		}
	}
	return suggestions
}

// errorRecovery checks spec.ErrorPatterns against envelope.Stderr.
func (g *Generic) errorRecovery(spec *toolspec.ToolSpec, envelope ContextEnvelope) *Suggestion {
	for _, ep := range spec.ErrorPatterns {
		if strings.Contains(envelope.Stderr, ep.Pattern) {
			return &Suggestion{
				Text:       ep.Fix,
				Confidence: MaxGenericConfidence,
				Primary:    true,
				Reason:     "error recovery",
			}
		}
	}
	return nil
}

// suggestFromCache uses the legacy Warm cache for backward compat.
func (g *Generic) suggestFromCache(envelope ContextEnvelope) *SuggestResponse {
	g.mu.RLock()
	subs, ok := g.cache[envelope.Cmd]
	g.mu.RUnlock()

	if !ok || len(subs) == 0 {
		return nil
	}

	suggestions := make([]Suggestion, 0, min(3, len(subs)))
	for i, sub := range subs {
		if i >= 3 {
			break
		}
		conf := MaxGenericConfidence - float64(i)*0.1
		if conf < 0.1 {
			conf = 0.1
		}
		suggestions = append(suggestions, Suggestion{
			Text:       envelope.Cmd + " " + sub,
			Confidence: conf,
			Primary:    i == 0,
		})
	}

	return &SuggestResponse{
		Intent:      envelope.Cmd,
		Stage:       "generic",
		Suggestions: suggestions,
	}
}

// subcommands returns cached subcommands or parses --help output.
// Deprecated: kept for backward compatibility; use Registry.Resolve instead.
func (g *Generic) subcommands(cmd string) []string {
	g.mu.RLock()
	if subs, ok := g.cache[cmd]; ok {
		g.mu.RUnlock()
		return subs
	}
	g.mu.RUnlock()

	subs := parseHelp(cmd)

	g.mu.Lock()
	g.cache[cmd] = subs
	g.mu.Unlock()

	return subs
}

// parseHelp runs `cmd --help` and extracts subcommand names.
// Deprecated: use kit/toolspec/sources/help instead.
func parseHelp(cmd string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	c := exec.CommandContext(ctx, cmd, "--help")
	out, err := c.CombinedOutput()
	if err != nil {
		return nil
	}

	return ParseHelpOutput(string(out))
}

// ParseHelpOutput extracts subcommand names from --help output text.
// Delegates to kit/toolspec/sources/help for the actual parsing.
// Kept exported for backward compatibility.
func ParseHelpOutput(text string) []string {
	spec := helpSource.ParseHelpOutput("", text)
	if spec == nil {
		return nil
	}
	names := make([]string, len(spec.Commands))
	for i, c := range spec.Commands {
		names[i] = c.Name
	}
	return names
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CacheHit reports whether cmd is in the help cache (for testing).
func (g *Generic) CacheHit(cmd string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, ok := g.cache[cmd]
	return ok
}

// Warm pre-populates the cache for testing.
func (g *Generic) Warm(cmd string, subs []string) {
	g.mu.Lock()
	g.cache[cmd] = subs
	g.mu.Unlock()
}

// SubcommandsTTL returns subcommands for a cmd, running --help with a
// custom timeout. Currently unused but reserved for future cache TTL.
func SubcommandsTTL(cmd string, timeout time.Duration) []string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c := exec.CommandContext(ctx, cmd, "--help")
	out, err := c.CombinedOutput()
	if err != nil {
		return nil
	}
	return ParseHelpOutput(string(out))
}
