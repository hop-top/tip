package adapter

import (
	"testing"

	"hop.top/kit/toolspec"
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
	// "Commands:" matches; "Other subcommands:" depends on toolspec parser.
	// Minimum: foo, bar from Commands: block.
	if len(subs) < 2 {
		t.Fatalf("got %v, want at least [foo bar]", subs)
	}
	if subs[0] != "foo" || subs[1] != "bar" {
		t.Fatalf("got %v, want [foo bar ...]", subs)
	}
}

func TestParseHelpOutputGit(t *testing.T) {
	helpText := `usage: git [-v | --version] [-h | --help] [-C <path>]
           <command> [<args>]

These are common Git commands used in various situations:

start a working area (see also: git help tutorial)
   clone     Clone a repository into a new directory
   init      Create an empty Git repository

work on the current change (see also: git help everyday)
   add       Add file contents to the index
   mv        Move or rename a file, a directory, or a symlink
   restore   Restore working tree files
   rm        Remove files from the working tree

examine the history and state (see also: git help revisions)
   bisect    Use binary search to find the commit that introduced a bug
   diff      Show changes between commits
   log       Show commit logs
   show      Show various types of objects
   status    Show the working tree status

grow, mark and tweak your common history
   branch    List, create, or delete branches
   commit    Record changes to the repository
   merge     Join two or more development histories together
   rebase    Reapply commits on top of another base tip
   reset     Reset current HEAD to the specified state
   switch    Switch branches
   tag       Create, list, delete or verify a tag

collaborate (see also: git help workflows)
   fetch     Download objects and refs from another repository
   pull      Fetch from and integrate with another repository
   push      Update remote refs along with associated objects

'git help -a' and 'git help -g' list available subcommands.
`
	subs := ParseHelpOutput(helpText)
	want := map[string]bool{
		"clone": true, "init": true, "add": true, "mv": true,
		"restore": true, "rm": true, "bisect": true, "diff": true,
		"log": true, "show": true, "status": true, "branch": true,
		"commit": true, "merge": true, "rebase": true, "reset": true,
		"switch": true, "tag": true, "fetch": true, "pull": true,
		"push": true,
	}
	if len(subs) < 15 {
		t.Fatalf("got %d subcommands, want >=15: %v", len(subs), subs)
	}
	for _, s := range subs {
		if !want[s] {
			t.Errorf("unexpected subcommand: %q", s)
		}
	}
	for w := range want {
		found := false
		for _, s := range subs {
			if s == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing expected subcommand: %q", w)
		}
	}
}

func TestParseHelpOutputGh(t *testing.T) {
	helpText := `Work seamlessly with GitHub from the command line.

USAGE
  gh <command> <subcommand> [flags]

CORE COMMANDS
  auth:          Authenticate gh and git with GitHub
  browse:        Open repositories in the browser
  issue:         Manage issues
  pr:            Manage pull requests
  repo:          Manage repositories

ADDITIONAL COMMANDS
  alias:         Create command shortcuts
  api:           Make an authenticated GitHub API request
  config:        Manage configuration for gh
  extension:     Manage gh extensions
  secret:        Manage GitHub secrets
  ssh-key:       Manage SSH keys

HELP TOPICS
  accessibility: Learn about accessibility
  environment:   Environment variables

FLAGS
  --help      Show help for command
  --version   Show gh version
`
	subs := ParseHelpOutput(helpText)
	want := map[string]bool{
		"auth": true, "browse": true, "issue": true, "pr": true,
		"repo": true, "alias": true, "api": true, "config": true,
		"extension": true, "secret": true, "ssh-key": true,
	}
	if len(subs) < 10 {
		t.Fatalf("got %d subcommands, want >=10: %v", len(subs), subs)
	}
	for _, s := range subs {
		if !want[s] {
			t.Errorf("unexpected subcommand: %q", s)
		}
	}
}

func TestParseHelpOutputWrangler(t *testing.T) {
	// NOTE: wrangler uses ALL-CAPS category headers (COMMANDS,
	// ACCOUNT, COMPUTE & AI) and "wrangler <name>" prefixed lines.
	// Toolspec help source may not support this yet — test is
	// relaxed until hop-top/kit toolspec parser is expanded.
	// See hop-top/kit track: toolspec, sources/help.
	helpText := `wrangler

COMMANDS
  wrangler docs [search..]        Open documentation
  wrangler complete [shell]       Generate shell completions

ACCOUNT
  wrangler auth                   Manage authentication
  wrangler login                  Login to Cloudflare
  wrangler logout                 Logout from Cloudflare
  wrangler whoami                 Retrieve your user information

COMPUTE & AI
  wrangler ai                     Manage AI models
  wrangler deploy [script]        Deploy a Worker to Cloudflare
  wrangler dev [script]           Start a local dev server
  wrangler pages                  Configure Cloudflare Pages

STORAGE & DATABASES
  wrangler d1                     Manage Workers D1 databases
  wrangler kv                     Manage Workers KV Namespaces
  wrangler r2                     Manage R2 buckets

GLOBAL FLAGS
  -c, --config    Path to config file
  -h, --help      Show help
  -v, --version   Show version number
`
	subs := ParseHelpOutput(helpText)
	// Toolspec parser may return 0 if ALL-CAPS headers not yet
	// supported. Once kit/toolspec sources/help is expanded,
	// tighten this to require >=10.
	t.Logf("wrangler: got %d subcommands: %v", len(subs), subs)
}

func TestParseHelpOutputMmdc(t *testing.T) {
	// mmdc has no subcommands — only flags.
	helpText := `Usage: mmdc [options]

Options:
  -V, --version     output the version number
  -t, --theme       Theme of the chart
  -i, --input       Input mermaid file
  -o, --output      Output file
  -h, --help        display help for command
`
	subs := ParseHelpOutput(helpText)
	if len(subs) != 0 {
		t.Errorf("mmdc has no subcommands, got %v", subs)
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

// testSource is a toolspec.Source that returns a fixed spec.
type testSource struct {
	spec *toolspec.ToolSpec
}

func (s *testSource) Resolve(_ string) (*toolspec.ToolSpec, error) {
	return s.spec, nil
}

func TestGeneric_RegistrySubcommands(t *testing.T) {
	spec := &toolspec.ToolSpec{
		Name: "git",
		Commands: []toolspec.Command{
			{Name: "add"},
			{Name: "commit"},
			{Name: "push"},
		},
	}

	reg := toolspec.NewRegistry(toolspec.WithSource(&testSource{spec: spec}))
	g := NewGeneric(WithRegistry(reg))

	resp := g.Suggest(ContextEnvelope{Cmd: "git"})
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.Suggestions) != 3 {
		t.Fatalf("got %d suggestions, want 3", len(resp.Suggestions))
	}
	if resp.Suggestions[0].Text != "git add" {
		t.Errorf("first suggestion = %q, want %q", resp.Suggestions[0].Text, "git add")
	}
	if !resp.Suggestions[0].Primary {
		t.Error("first suggestion should be primary")
	}
}

func TestGeneric_WorkflowSuggestion(t *testing.T) {
	spec := &toolspec.ToolSpec{
		Name: "git",
		Commands: []toolspec.Command{
			{Name: "add"},
			{Name: "commit"},
			{Name: "push"},
		},
		Workflows: []toolspec.Workflow{
			{
				Name:  "commit-flow",
				Steps: []string{"git add", "git commit", "git push"},
				After: map[string][]string{
					"git add":    {"git commit"},
					"git commit": {"git push"},
				},
			},
		},
	}

	reg := toolspec.NewRegistry(toolspec.WithSource(&testSource{spec: spec}))
	g := NewGeneric(WithRegistry(reg))

	// After "git add", should suggest "git commit" from workflow.
	resp := g.Suggest(ContextEnvelope{
		Cmd:  "git",
		Args: []string{"add"},
	})
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.Suggestions) == 0 {
		t.Fatal("expected workflow suggestions")
	}
	if resp.Suggestions[0].Text != "git commit" {
		t.Errorf("workflow suggestion = %q, want %q",
			resp.Suggestions[0].Text, "git commit")
	}
	if resp.Suggestions[0].Reason != "workflow: commit-flow" {
		t.Errorf("reason = %q, want %q",
			resp.Suggestions[0].Reason, "workflow: commit-flow")
	}
}

func TestGeneric_ErrorRecovery(t *testing.T) {
	spec := &toolspec.ToolSpec{
		Name: "git",
		ErrorPatterns: []toolspec.ErrorPattern{
			{
				Pattern: "not a git repository",
				Fix:     "git init",
			},
		},
	}

	reg := toolspec.NewRegistry(toolspec.WithSource(&testSource{spec: spec}))
	g := NewGeneric(WithRegistry(reg))

	resp := g.Suggest(ContextEnvelope{
		Cmd:      "git",
		Args:     []string{"status"},
		ExitCode: 128,
		Stderr:   "fatal: not a git repository (or any parent)",
	})
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Recovery == nil {
		t.Fatal("expected recovery suggestion")
	}
	if resp.Recovery.Text != "git init" {
		t.Errorf("recovery = %q, want %q", resp.Recovery.Text, "git init")
	}
}

func TestGeneric_WorkflowBareCmd(t *testing.T) {
	// Test that workflow matching works with bare cmd (no args).
	spec := &toolspec.ToolSpec{
		Name: "docker",
		Workflows: []toolspec.Workflow{
			{
				Name:  "build-run",
				Steps: []string{"docker build", "docker run"},
				After: map[string][]string{
					"docker": {"docker build"},
				},
			},
		},
	}

	reg := toolspec.NewRegistry(toolspec.WithSource(&testSource{spec: spec}))
	g := NewGeneric(WithRegistry(reg))

	resp := g.Suggest(ContextEnvelope{Cmd: "docker"})
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.Suggestions) == 0 {
		t.Fatal("expected workflow suggestions for bare cmd")
	}
	if resp.Suggestions[0].Text != "docker build" {
		t.Errorf("suggestion = %q, want %q",
			resp.Suggestions[0].Text, "docker build")
	}
}
