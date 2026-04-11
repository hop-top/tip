// Package render formats tip suggestions for terminal output using lipgloss.
//
// Rules:
//   - 1 primary line (arrow prefix), up to 2 secondary (indented, dim),
//     and 1 recovery (fix:/try: prefix, warm color)
//   - Max 4 lines total
//   - Respects NO_COLOR env var
package render

import (
	"os"
	"strings"

	"charm.land/lipgloss/v2"
)

// Suggestion is a single next-step suggestion.
type Suggestion struct {
	Text    string
	Primary bool
}

// Recovery holds error-recovery suggestions.
type Recovery struct {
	Fix string
	Try string
}

// Result is the full set of suggestions to render.
type Result struct {
	Suggestions []Suggestion
	Recovery    *Recovery
}

// Options controls rendering behavior.
type Options struct {
	NoColor bool
}

// Format renders the result to a string.
func Format(r Result, opts Options) string {
	if len(r.Suggestions) == 0 && r.Recovery == nil {
		return ""
	}

	noColor := opts.NoColor || os.Getenv("NO_COLOR") != ""

	var lines []string

	// Collect primary
	for _, s := range r.Suggestions {
		if s.Primary {
			lines = append(lines, formatPrimary(s.Text, noColor))
			break
		}
	}

	// Collect secondary (max 2)
	sec := 0
	for _, s := range r.Suggestions {
		if s.Primary {
			continue
		}
		if sec >= 2 {
			break
		}
		lines = append(lines, formatSecondary(s.Text, noColor))
		sec++
	}

	// Recovery (max 1 line, prefer fix over try)
	if r.Recovery != nil && len(lines) < maxLines {
		if r.Recovery.Fix != "" {
			lines = append(lines, formatRecovery("fix", r.Recovery.Fix, noColor))
		} else if r.Recovery.Try != "" {
			lines = append(lines, formatRecovery("try", r.Recovery.Try, noColor))
		}
		// Add try if we still have room and fix was used
		if r.Recovery.Fix != "" && r.Recovery.Try != "" && len(lines) < maxLines {
			lines = append(lines, formatRecovery("try", r.Recovery.Try, noColor))
		}
	}

	// Cap at maxLines
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	return strings.Join(lines, "\n") + "\n"
}

const maxLines = 4

func formatPrimary(text string, noColor bool) string {
	if noColor {
		return arrowPlain + text
	}
	arrow := lipgloss.NewStyle().Faint(true).Render(arrowPlain)
	return arrow + text
}

func formatSecondary(text string, noColor bool) string {
	if noColor {
		return "  " + text
	}
	return lipgloss.NewStyle().Faint(true).Render("  " + text)
}

func formatRecovery(kind, text string, noColor bool) string {
	prefix := "  " + kind + ": "
	if noColor {
		return prefix + text
	}
	amber := lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#D4A017"))
	return amber.Render(prefix + text)
}

const arrowPlain = "\u2192 " // → + space
