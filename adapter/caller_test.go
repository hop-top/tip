package adapter

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// buildHelper compiles a small Go program in dir and returns the binary path.
func buildHelper(t *testing.T, dir, name, src string) string {
	t.Helper()

	srcFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(srcFile, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	bin := filepath.Join(dir, name)
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", bin, srcFile)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build %s: %v\n%s", name, err, out)
	}
	return bin
}

func TestCallSuggest(t *testing.T) {
	dir := t.TempDir()
	bin := buildHelper(t, dir, "tip-echo", `
package main

import (
	"encoding/json"
	"io"
	"os"
)

type Envelope struct {
	Cmd  string   `+"`"+`json:"cmd"`+"`"+`
	Args []string `+"`"+`json:"args"`+"`"+`
}

type Suggestion struct {
	Text       string  `+"`"+`json:"text"`+"`"+`
	Confidence float64 `+"`"+`json:"confidence"`+"`"+`
	Primary    bool    `+"`"+`json:"primary,omitempty"`+"`"+`
}

type Response struct {
	Intent      string       `+"`"+`json:"intent"`+"`"+`
	Stage       string       `+"`"+`json:"stage"`+"`"+`
	Suggestions []Suggestion `+"`"+`json:"suggestions"`+"`"+`
}

func main() {
	data, _ := io.ReadAll(os.Stdin)
	var env Envelope
	json.Unmarshal(data, &env)

	resp := Response{
		Intent: env.Cmd,
		Stage:  "post",
		Suggestions: []Suggestion{
			{Text: env.Cmd + " push", Confidence: 0.9, Primary: true},
		},
	}
	json.NewEncoder(os.Stdout).Encode(resp)
}
`)

	env := ContextEnvelope{Cmd: "git", Args: []string{"commit"}}
	resp, err := Call(bin, env)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Intent != "git" {
		t.Errorf("Intent = %q, want git", resp.Intent)
	}
	if len(resp.Suggestions) != 1 {
		t.Fatalf("Suggestions len = %d, want 1", len(resp.Suggestions))
	}
	if resp.Suggestions[0].Text != "git push" {
		t.Errorf("Suggestion text = %q, want %q", resp.Suggestions[0].Text, "git push")
	}
}

func TestCallTimeout(t *testing.T) {
	dir := t.TempDir()
	bin := buildHelper(t, dir, "tip-slow", `
package main

import "time"

func main() {
	time.Sleep(5 * time.Second)
}
`)

	env := ContextEnvelope{Cmd: "slow"}
	resp, err := CallWithTimeout(bin, env, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if resp != nil {
		t.Error("expected nil response on timeout")
	}
}

func TestCallNonZeroExit(t *testing.T) {
	dir := t.TempDir()
	bin := buildHelper(t, dir, "tip-fail", `
package main

import "os"

func main() {
	os.Exit(1)
}
`)

	env := ContextEnvelope{Cmd: "fail"}
	resp, err := Call(bin, env)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if resp != nil {
		t.Error("expected nil response on non-zero exit")
	}
}
