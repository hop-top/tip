package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"hop.top/tip/session"
)

func TestBuildEnvelope(t *testing.T) {
	parsed := Parsed{Cmd: "git", Args: []string{"push"}}
	history := []session.Command{
		{Cmd: "git", Args: []string{"add", "."}, ExitCode: 0},
		{Cmd: "git", Args: []string{"commit", "-m", "wip"}, ExitCode: 0},
	}

	env := Build(parsed, 0, "", "/tmp/proj", history)

	assert.Equal(t, "git", env.Cmd)
	assert.Equal(t, []string{"push"}, env.Args)
	assert.Equal(t, 0, env.ExitCode)
	assert.Equal(t, "/tmp/proj", env.Cwd)
	assert.Len(t, env.History, 2)
	assert.Equal(t, "git", env.History[0].Cmd)
	assert.Equal(t, []string{"add", "."}, env.History[0].Args)
}

func TestBuildEnvelopeEmptyHistory(t *testing.T) {
	parsed := Parsed{Cmd: "ls", Args: nil}
	env := Build(parsed, 0, "", "/tmp", nil)

	assert.Equal(t, "ls", env.Cmd)
	assert.Empty(t, env.History)
}

func TestBuildEnvelopeWithStderr(t *testing.T) {
	parsed := Parsed{Cmd: "docker", Args: []string{"build", "."}}
	env := Build(parsed, 1, "error: no Dockerfile", "/tmp", nil)

	assert.Equal(t, 1, env.ExitCode)
	assert.Equal(t, "error: no Dockerfile", env.Stderr)
}
