// Package policy ranks, filters, and caps adapter suggestions.
//
// Pipeline: dedup -> cap generic -> filter unsafe -> rank -> limit.
package policy

import (
	"sort"
	"strings"
)

// Suggestion is a single next-step suggestion from an adapter.
type Suggestion struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Primary    bool    `json:"primary,omitempty"`
	Reason     string  `json:"reason,omitempty"`
	Recovery   bool    `json:"recovery,omitempty"`
}

// Result holds the filtered and ranked suggestions after policy.
type Result struct {
	Primary    *Suggestion  // At most one primary.
	Secondary  []Suggestion // Up to 2 secondary.
	Recovery   *Suggestion  // At most one recovery.
}

// Options controls policy behaviour.
type Options struct {
	// IsGeneric is true when suggestions came from the generic adapter.
	IsGeneric bool
	// OriginalCmd is the command the user ran, used to decide whether
	// destructive suggestions should be filtered.
	OriginalCmd string
}

// genericCap is the max confidence for generic adapter suggestions.
const genericCap = 0.6

// unsafePatterns that are filtered unless the original command was
// already destructive.
var unsafePatterns = []string{
	"rm -rf",
	"--force",
	"reset --hard",
}

// destructivePatterns in the original command that exempt unsafe
// suggestion filtering.
var destructivePatterns = []string{
	"rm ",
	"--force",
	"reset --hard",
}

// Apply runs the full policy pipeline on suggestions.
func Apply(suggestions []Suggestion, opts Options) Result {
	s := dedup(suggestions)
	s = capGeneric(s, opts.IsGeneric)
	s = filterUnsafe(s, opts.OriginalCmd)
	s = rank(s)
	return limit(s)
}

// dedup removes duplicate suggestion text, keeping highest confidence.
func dedup(suggestions []Suggestion) []Suggestion {
	seen := make(map[string]int) // text -> index in result
	var result []Suggestion
	for _, s := range suggestions {
		key := strings.TrimSpace(s.Text)
		s.Text = key
		if idx, ok := seen[key]; ok {
			if s.Confidence > result[idx].Confidence {
				result[idx] = s
			}
			continue
		}
		seen[key] = len(result)
		result = append(result, s)
	}
	return result
}

// capGeneric caps confidence at genericCap for generic adapter.
func capGeneric(suggestions []Suggestion, isGeneric bool) []Suggestion {
	if !isGeneric {
		return suggestions
	}
	result := make([]Suggestion, len(suggestions))
	for i, s := range suggestions {
		result[i] = s
		if s.Confidence > genericCap {
			result[i].Confidence = genericCap
		}
	}
	return result
}

// filterUnsafe removes destructive suggestions unless the original
// command was already destructive.
func filterUnsafe(suggestions []Suggestion, originalCmd string) []Suggestion {
	if isDestructive(originalCmd) {
		return suggestions
	}
	var result []Suggestion
	for _, s := range suggestions {
		if isUnsafe(s.Text) {
			continue
		}
		result = append(result, s)
	}
	return result
}

func isUnsafe(text string) bool {
	lower := strings.ToLower(text)
	for _, p := range unsafePatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func isDestructive(cmd string) bool {
	lower := strings.ToLower(cmd)
	for _, p := range destructivePatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// rank sorts by confidence descending.
func rank(suggestions []Suggestion) []Suggestion {
	result := make([]Suggestion, len(suggestions))
	copy(result, suggestions)
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Confidence > result[j].Confidence
	})
	return result
}

// limit caps output to 1 primary + 2 secondary + 1 recovery.
func limit(suggestions []Suggestion) Result {
	var res Result
	for _, s := range suggestions {
		if s.Recovery {
			if res.Recovery == nil {
				v := s
				res.Recovery = &v
			}
			continue
		}
		if s.Primary && res.Primary == nil {
			v := s
			res.Primary = &v
			continue
		}
		if res.Primary == nil && len(res.Secondary) == 0 {
			// First non-recovery becomes primary if none marked.
			v := s
			v.Primary = true
			res.Primary = &v
			continue
		}
		if len(res.Secondary) < 2 {
			s.Primary = false
			res.Secondary = append(res.Secondary, s)
		}
	}
	return res
}
