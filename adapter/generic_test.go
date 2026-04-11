package adapter

import (
	"testing"
)

func TestParseHelpOutput(t *testing.T) {
	helpText := `Usage: mycli <command> [flags]

Available Commands:
  init        Initialize a new project
  build       Build the project
  test        Run tests
  deploy      Deploy to production

Flags:
  -h, --help   Show help
  -v, --version  Show version
`

	subs := ParseHelpOutput(helpText)
	want := []string{"init", "build", "test", "deploy"}
	if len(subs) != len(want) {
		t.Fatalf("got %d subcommands, want %d: %v", len(subs), len(want), subs)
	}
	for i, s := range subs {
		if s != want[i] {
			t.Errorf("subs[%d] = %q, want %q", i, s, want[i])
		}
	}
}

func TestParseHelpOutputMultipleBlocks(t *testing.T) {
	helpText := `Commands:
  foo   Do foo
  bar   Do bar

Other subcommands:
  baz   Do baz
`
	subs := ParseHelpOutput(helpText)
	if len(subs) != 2 {
		// "Other subcommands:" ends with "subcommands:" so baz is included
		// Actually our parser checks HasSuffix "commands:" which matches
		// "subcommands:" too. Let's verify.
	}
	// "Commands:" -> foo, bar. "Other subcommands:" -> baz
	want := []string{"foo", "bar", "baz"}
	if len(subs) != len(want) {
		t.Fatalf("got %v, want %v", subs, want)
	}
}

func TestParseHelpOutputEmpty(t *testing.T) {
	subs := ParseHelpOutput("Usage: tool [flags]\n  -h  help\n")
	if len(subs) != 0 {
		t.Errorf("expected no subcommands, got %v", subs)
	}
}

func TestGenericSuggestCappedConfidence(t *testing.T) {
	g := NewGeneric()
	g.Warm("mytool", []string{"init", "build", "test", "deploy"})

	env := ContextEnvelope{Cmd: "mytool", Args: []string{"build"}, ExitCode: 0}
	resp := g.Suggest(env)
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	for _, s := range resp.Suggestions {
		if s.Confidence > MaxGenericConfidence {
			t.Errorf("confidence %f exceeds cap %f for %q",
				s.Confidence, MaxGenericConfidence, s.Text)
		}
	}

	if len(resp.Suggestions) > 3 {
		t.Errorf("suggestions count = %d, want <= 3", len(resp.Suggestions))
	}
}

func TestGenericCacheHit(t *testing.T) {
	g := NewGeneric()

	if g.CacheHit("testcmd") {
		t.Error("cache should miss before Warm")
	}

	g.Warm("testcmd", []string{"sub1"})

	if !g.CacheHit("testcmd") {
		t.Error("cache should hit after Warm")
	}

	// Second call should use cache
	resp1 := g.Suggest(ContextEnvelope{Cmd: "testcmd"})
	resp2 := g.Suggest(ContextEnvelope{Cmd: "testcmd"})
	if resp1 == nil || resp2 == nil {
		t.Fatal("expected non-nil responses")
	}
	if resp1.Suggestions[0].Text != resp2.Suggestions[0].Text {
		t.Error("cache should return same results")
	}
}

func TestGenericSuggestNoSubcommands(t *testing.T) {
	g := NewGeneric()
	g.Warm("empty", []string{})

	resp := g.Suggest(ContextEnvelope{Cmd: "empty"})
	if resp != nil {
		t.Error("expected nil response for empty subcommands")
	}
}

func TestGenericSuggestPrimary(t *testing.T) {
	g := NewGeneric()
	g.Warm("tool", []string{"a", "b", "c"})

	resp := g.Suggest(ContextEnvelope{Cmd: "tool"})
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if !resp.Suggestions[0].Primary {
		t.Error("first suggestion should be primary")
	}
	for _, s := range resp.Suggestions[1:] {
		if s.Primary {
			t.Errorf("non-first suggestion %q should not be primary", s.Text)
		}
	}
}
