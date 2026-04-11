package render_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"hop.top/tip/render"
)

func TestSuccessFormat(t *testing.T) {
	result := render.Result{
		Suggestions: []render.Suggestion{
			{Text: "git push", Primary: true},
			{Text: "git log --oneline -5"},
		},
	}
	out := render.Format(result, render.Options{NoColor: true})

	assert.Contains(t, out, "git push")
	assert.Contains(t, out, "git log --oneline -5")
	// Primary gets arrow prefix
	lines := strings.Split(strings.TrimSpace(out), "\n")
	assert.True(t, len(lines) >= 2)
}

func TestRecoveryFormat(t *testing.T) {
	result := render.Result{
		Recovery: &render.Recovery{
			Fix: "docker login registry",
			Try: "docker build --no-cache",
		},
	}
	out := render.Format(result, render.Options{NoColor: true})

	assert.Contains(t, out, "fix:")
	assert.Contains(t, out, "docker login registry")
	assert.Contains(t, out, "try:")
	assert.Contains(t, out, "docker build --no-cache")
}

func TestNoColorStripsANSI(t *testing.T) {
	result := render.Result{
		Suggestions: []render.Suggestion{
			{Text: "git push", Primary: true},
		},
	}
	out := render.Format(result, render.Options{NoColor: true})
	assert.NotContains(t, out, "\x1b[")
}

func TestMaxLinesEnforced(t *testing.T) {
	result := render.Result{
		Suggestions: []render.Suggestion{
			{Text: "line1", Primary: true},
			{Text: "line2"},
			{Text: "line3"},
			{Text: "line4"},
			{Text: "line5"},
			{Text: "line6"},
		},
		Recovery: &render.Recovery{
			Fix: "fix something",
		},
	}
	out := render.Format(result, render.Options{NoColor: true})
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// Max 4 lines: 1 primary + 2 secondary + 1 recovery
	assert.LessOrEqual(t, len(lines), 4)
}

func TestEmptyResult(t *testing.T) {
	out := render.Format(render.Result{}, render.Options{NoColor: true})
	assert.Equal(t, "", out)
}

func TestColorOutput(t *testing.T) {
	result := render.Result{
		Suggestions: []render.Suggestion{
			{Text: "git push", Primary: true},
		},
	}
	out := render.Format(result, render.Options{NoColor: false})
	// With color enabled, there should be styling
	assert.Contains(t, out, "git push")
}
