package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSimple(t *testing.T) {
	p := Parse("git push")
	assert.Equal(t, "git", p.Cmd)
	assert.Equal(t, []string{"push"}, p.Args)
}

func TestParseWithFlags(t *testing.T) {
	p := Parse("docker build -t myapp .")
	assert.Equal(t, "docker", p.Cmd)
	assert.Equal(t, []string{"build", "-t", "myapp", "."}, p.Args)
}

func TestParseChainedAnd(t *testing.T) {
	p := Parse("git add . && git commit -m 'wip'")
	assert.Equal(t, "git", p.Cmd)
	assert.Equal(t, []string{"commit", "-m", "wip"}, p.Args)
}

func TestParseChainedSemicolon(t *testing.T) {
	p := Parse("echo hello; git status")
	assert.Equal(t, "git", p.Cmd)
	assert.Equal(t, []string{"status"}, p.Args)
}

func TestParsePipe(t *testing.T) {
	p := Parse("cat file.txt | grep error")
	assert.Equal(t, "grep", p.Cmd)
	assert.Equal(t, []string{"error"}, p.Args)
}

func TestParseEmpty(t *testing.T) {
	p := Parse("")
	assert.Empty(t, p.Cmd)
	assert.Nil(t, p.Args)
}

func TestParseQuotedArgs(t *testing.T) {
	p := Parse(`git commit -m "fix: the thing"`)
	assert.Equal(t, "git", p.Cmd)
	assert.Equal(t, []string{"commit", "-m", "fix: the thing"}, p.Args)
}

func TestParseNoArgs(t *testing.T) {
	p := Parse("ls")
	assert.Equal(t, "ls", p.Cmd)
	assert.Empty(t, p.Args)
}

func TestParseMixedChain(t *testing.T) {
	p := Parse("make build && ./bin/app | tee log.txt")
	assert.Equal(t, "tee", p.Cmd)
	assert.Equal(t, []string{"log.txt"}, p.Args)
}
