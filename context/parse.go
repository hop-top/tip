package context

import (
	"strings"
)

// Parsed holds the result of parsing a raw command line.
type Parsed struct {
	Cmd  string   // The base command (last in a chain).
	Args []string // Arguments to the command.
}

// Parse splits a raw command line into cmd + args.
//
// For chained commands (&&, ;, |), the last command in the chain is
// used since that is the one whose exit code the shell reports.
func Parse(raw string) Parsed {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Parsed{}
	}

	// Take the last command in a chain.
	last := lastInChain(raw)
	last = strings.TrimSpace(last)
	if last == "" {
		return Parsed{}
	}

	tokens := tokenize(last)
	if len(tokens) == 0 {
		return Parsed{}
	}
	return Parsed{
		Cmd:  tokens[0],
		Args: tokens[1:],
	}
}

// lastInChain returns the last segment after splitting on &&, ;, or |.
// Pipe is included because the exit code comes from the last command
// in the pipeline. Separators inside single or double quotes are ignored.
func lastInChain(s string) string {
	var inSingle, inDouble bool
	lastSep := -1 // index past the last unquoted separator
	sepLen := 0

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch == '\'' && !inDouble:
			inSingle = !inSingle
		case ch == '"' && !inSingle:
			inDouble = !inDouble
		case !inSingle && !inDouble:
			if ch == ';' {
				lastSep = i
				sepLen = 1
			} else if ch == '|' && (i == 0 || s[i-1] != '|') &&
				(i+1 >= len(s) || s[i+1] != '|') {
				lastSep = i
				sepLen = 1
			} else if ch == '&' && i+1 < len(s) && s[i+1] == '&' {
				lastSep = i
				sepLen = 2
				i++ // skip second '&'
			}
		}
	}
	if lastSep < 0 {
		return s
	}
	return s[lastSep+sepLen:]
}

// tokenize performs a simple shell-like split respecting single and
// double quotes. It does not handle escapes or variable expansion.
func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder
	var inSingle, inDouble bool

	for _, r := range s {
		switch {
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case r == ' ' && !inSingle && !inDouble:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}
