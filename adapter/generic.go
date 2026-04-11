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
// Handles multiple CLI frameworks:
//   - Cobra/Docker: "Commands:", "Available Commands:", etc.
//   - gh (cli/cli): "CORE COMMANDS", "ADDITIONAL COMMANDS" (ALL-CAPS, no colon)
//   - wrangler (yargs): "COMMANDS", "COMPUTE & AI" (ALL-CAPS categories)
//   - git: prose category headers with indented subcommands
//
// Exported for testing.
func ParseHelpOutput(text string) []string {
	var (
		subs    []string
		seen    = map[string]bool{}
		inBlock bool
	)

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		// Check if this line is a section header that precedes
		// subcommand listings. Multiple formats supported:

		if isSectionHeader(trimmed, lower) {
			inBlock = true
			continue
		}

		// Non-section-header, non-indented line ends the block
		// (unless it's empty, which also ends it).
		if inBlock {
			if trimmed == "" {
				inBlock = false
				continue
			}

			// Must be indented (leading whitespace) to be a
			// subcommand line — non-indented = new section or
			// end of block.
			if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
				// Could be the next section header.
				if isSectionHeader(trimmed, lower) {
					continue // stay in block mode
				}
				inBlock = false
				continue
			}

			name := extractSubcommand(trimmed)
			if name != "" && !seen[name] {
				seen[name] = true
				subs = append(subs, name)
			}
		}
	}
	if scanner.Err() != nil {
		return nil
	}

	return subs
}

// isSectionHeader checks if a line is a header that precedes
// a subcommand listing.
func isSectionHeader(trimmed, lower string) bool {
	// Pattern 1: ends with "commands:" or "subcommands:"
	// Matches: "Available Commands:", "Common Commands:",
	//          "Management Commands:", "Other subcommands:"
	if strings.HasSuffix(lower, "commands:") ||
		strings.HasSuffix(lower, "subcommands:") {
		return true
	}

	// Pattern 2: ALL-CAPS line that is likely a category header.
	// Matches: "CORE COMMANDS", "COMMANDS", "COMPUTE & AI",
	//          "STORAGE & DATABASES", "ACCOUNT"
	// Excludes: "FLAGS", "OPTIONS", "USAGE", "HELP TOPICS",
	//           "GLOBAL FLAGS", "EXAMPLES", "LEARN MORE"
	if isAllCaps(trimmed) && len(trimmed) > 2 {
		if isNonCommandSection(lower) {
			return false
		}
		return true
	}

	// Pattern 3: git-style prose category header.
	// Lines like "start a working area (see also: ...)" followed
	// by 3-space-indented subcommands. Detected by checking if
	// the line is not indented, contains lowercase words, and is
	// followed by indented content (handled by the caller staying
	// in block mode).
	if strings.Contains(lower, "(see also:") ||
		strings.Contains(lower, "these are common") {
		return true
	}

	// Pattern 4: git category without parenthetical.
	// Lines like "grow, mark and tweak your common history"
	// or "collaborate (see also: git help workflows)"
	// Heuristic: non-indented, all lowercase, no flag-like
	// chars, and not a usage/description line.
	if isGitCategory(trimmed, lower) {
		return true
	}

	return false
}

// isAllCaps checks if a string is all uppercase letters, spaces,
// ampersands, and punctuation (no lowercase).
func isAllCaps(s string) bool {
	hasLetter := false
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return false
		}
		if r >= 'A' && r <= 'Z' {
			hasLetter = true
		}
	}
	return hasLetter
}

// isNonCommandSection returns true for ALL-CAPS headers that
// list flags, options, or help topics rather than subcommands.
var skipSections = map[string]bool{
	"flags": true, "global flags": true, "global options": true,
	"options": true, "usage": true, "examples": true,
	"help topics": true, "learn more": true, "resources": true,
	"description": true, "synopsis": true, "environment": true,
	"see also": true, "diagnostics": true,
}

func isNonCommandSection(lower string) bool {
	return skipSections[lower]
}

// isGitCategory detects git-style prose category headers.
// These are non-indented, all-lowercase lines that describe
// a group of subcommands.
func isGitCategory(trimmed, lower string) bool {
	if trimmed == "" || trimmed != lower {
		return false // must be all lowercase
	}
	if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "'") {
		return false // flags or quotes
	}
	// Must contain a comma or "and" — git categories are phrases
	// like "grow, mark and tweak your common history"
	if !strings.Contains(lower, ",") && !strings.Contains(lower, " and ") {
		return false
	}
	return true
}

// extractSubcommand extracts the subcommand name from an indented line.
// Handles multiple formats:
//   - "  name        Description..."         (docker, cobra)
//   - "  name:       Description..."         (gh)
//   - "  prefix name [args]  📚 Description" (wrangler)
//   - "   name     Description..."           (git, 3-space indent)
func extractSubcommand(line string) string {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return ""
	}

	name := parts[0]

	// Skip lines that start with flags.
	if strings.HasPrefix(name, "-") {
		return ""
	}

	// wrangler format: "wrangler <subcommand> [args]..."
	// The actual subcommand is parts[1].
	if len(parts) > 1 && (name == "wrangler" || name == "gh" ||
		name == "docker" || name == "git") {
		name = parts[1]
	}

	// gh format: "name:" — strip trailing colon.
	name = strings.TrimSuffix(name, ":")

	// Skip names with brackets (args like "[search..]").
	if strings.HasPrefix(name, "[") || strings.HasPrefix(name, "<") {
		return ""
	}

	// Skip emoji-only tokens.
	if len(name) > 0 && name[0] > 127 {
		return ""
	}

	// Skip names that look like decorators or noise.
	if strings.Contains(name, "*") {
		name = strings.TrimSuffix(name, "*")
	}

	if name == "" {
		return ""
	}

	return name
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
