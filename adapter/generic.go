package adapter

import (
	"bufio"
	"context"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// MaxGenericConfidence caps confidence for generic adapter suggestions.
const MaxGenericConfidence = 0.6

// Generic is the built-in Level 0 adapter that works with any CLI tool.
// It parses `<cmd> --help` output to extract subcommands and returns
// suggestions with confidence capped at MaxGenericConfidence.
type Generic struct {
	mu    sync.RWMutex
	cache map[string][]string // cmd -> subcommands
}

// NewGeneric creates a new Generic adapter with an empty cache.
func NewGeneric() *Generic {
	return &Generic{cache: make(map[string][]string)}
}

// Suggest returns suggestions based on --help output of the command.
func (g *Generic) Suggest(envelope ContextEnvelope) *SuggestResponse {
	subs := g.subcommands(envelope.Cmd)
	if len(subs) == 0 {
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
// Exported for testing.
func ParseHelpOutput(text string) []string {
	var (
		subs    []string
		inBlock bool
	)

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		lower := strings.ToLower(strings.TrimSpace(line))

		// Detect "commands:", "available commands:", etc.
		if strings.HasSuffix(lower, "commands:") ||
			strings.HasSuffix(lower, "subcommands:") {
			inBlock = true
			continue
		}

		if inBlock {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				inBlock = false
				continue
			}
			// First word is the subcommand name.
			parts := strings.Fields(trimmed)
			if len(parts) > 0 && !strings.HasPrefix(parts[0], "-") {
				subs = append(subs, parts[0])
			}
		}
	}

	return subs
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
