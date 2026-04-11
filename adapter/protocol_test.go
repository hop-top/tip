package adapter

import (
	"encoding/json"
	"testing"
)

func TestDescribeResponseRoundTrip(t *testing.T) {
	orig := DescribeResponse{
		Name:           "git",
		Version:        "0.0.1",
		ToolVersionCmd: "git --version",
		Commands:       []string{"commit", "push", "pull"},
		Capabilities:   []string{"workflows", "error_patterns"},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got DescribeResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Name != orig.Name {
		t.Errorf("Name = %q, want %q", got.Name, orig.Name)
	}
	if got.Version != orig.Version {
		t.Errorf("Version = %q, want %q", got.Version, orig.Version)
	}
	if got.ToolVersionCmd != orig.ToolVersionCmd {
		t.Errorf("ToolVersionCmd = %q, want %q", got.ToolVersionCmd, orig.ToolVersionCmd)
	}
	if len(got.Commands) != len(orig.Commands) {
		t.Errorf("Commands len = %d, want %d", len(got.Commands), len(orig.Commands))
	}
	if len(got.Capabilities) != len(orig.Capabilities) {
		t.Errorf("Capabilities len = %d, want %d", len(got.Capabilities), len(orig.Capabilities))
	}
}

func TestSuggestResponseRoundTrip(t *testing.T) {
	orig := SuggestResponse{
		Intent: "commit",
		Stage:  "post-commit",
		Suggestions: []Suggestion{
			{Text: "git push", Confidence: 0.9, Primary: true, Reason: "unpushed commits"},
			{Text: "git log --oneline -5", Confidence: 0.5},
		},
		Recovery: nil,
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got SuggestResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Intent != orig.Intent {
		t.Errorf("Intent = %q, want %q", got.Intent, orig.Intent)
	}
	if got.Stage != orig.Stage {
		t.Errorf("Stage = %q, want %q", got.Stage, orig.Stage)
	}
	if len(got.Suggestions) != 2 {
		t.Fatalf("Suggestions len = %d, want 2", len(got.Suggestions))
	}
	if !got.Suggestions[0].Primary {
		t.Error("Suggestions[0].Primary should be true")
	}
	if got.Suggestions[1].Primary {
		t.Error("Suggestions[1].Primary should be false")
	}
}

func TestSuggestResponseWithRecovery(t *testing.T) {
	orig := SuggestResponse{
		Intent:      "build",
		Stage:       "post-error",
		Suggestions: []Suggestion{},
		Recovery:    &Suggestion{Text: "docker login", Confidence: 0.8, Reason: "auth required"},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got SuggestResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Recovery == nil {
		t.Fatal("Recovery should not be nil")
	}
	if got.Recovery.Text != "docker login" {
		t.Errorf("Recovery.Text = %q, want %q", got.Recovery.Text, "docker login")
	}
}

func TestContextEnvelopeRoundTrip(t *testing.T) {
	orig := ContextEnvelope{
		Cmd:      "git",
		Args:     []string{"commit", "-m", "fix: typo"},
		ExitCode: 0,
		Stderr:   "",
		Cwd:      "/home/user/project",
		History: []HistoryEntry{
			{Cmd: "git", Args: []string{"add", "."}, ExitCode: 0},
		},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got ContextEnvelope
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Cmd != orig.Cmd {
		t.Errorf("Cmd = %q, want %q", got.Cmd, orig.Cmd)
	}
	if len(got.Args) != 3 {
		t.Errorf("Args len = %d, want 3", len(got.Args))
	}
	if got.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", got.ExitCode)
	}
	if got.Cwd != orig.Cwd {
		t.Errorf("Cwd = %q, want %q", got.Cwd, orig.Cwd)
	}
	if len(got.History) != 1 {
		t.Fatalf("History len = %d, want 1", len(got.History))
	}
	if got.History[0].Cmd != "git" {
		t.Errorf("History[0].Cmd = %q, want %q", got.History[0].Cmd, "git")
	}
}

func TestDescribeResponseJSON(t *testing.T) {
	// Verify JSON field names match the protocol spec.
	raw := `{
		"name": "docker",
		"version": "0.1.0",
		"tool_version_cmd": "docker --version",
		"commands": ["build", "run"],
		"capabilities": ["error_patterns"]
	}`

	var d DescribeResponse
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		t.Fatalf("unmarshal from spec JSON: %v", err)
	}

	if d.Name != "docker" {
		t.Errorf("Name = %q, want docker", d.Name)
	}
	if d.ToolVersionCmd != "docker --version" {
		t.Errorf("ToolVersionCmd = %q", d.ToolVersionCmd)
	}
	if len(d.Commands) != 2 {
		t.Errorf("Commands len = %d, want 2", len(d.Commands))
	}
}
