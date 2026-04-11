package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDedupKeepsHighestConfidence(t *testing.T) {
	in := []Suggestion{
		{Text: "git push", Confidence: 0.5},
		{Text: "git push", Confidence: 0.9},
		{Text: "git log", Confidence: 0.3},
	}
	out := dedup(in)
	require.Len(t, out, 2)
	assert.Equal(t, 0.9, out[0].Confidence)
	assert.Equal(t, "git log", out[1].Text)
}

func TestCapGeneric(t *testing.T) {
	in := []Suggestion{
		{Text: "git push", Confidence: 0.9},
		{Text: "git log", Confidence: 0.4},
	}
	out := capGeneric(in, true)
	assert.Equal(t, 0.6, out[0].Confidence)
	assert.Equal(t, 0.4, out[1].Confidence)
}

func TestCapGenericNoOpWhenNotGeneric(t *testing.T) {
	in := []Suggestion{
		{Text: "git push", Confidence: 0.9},
	}
	out := capGeneric(in, false)
	assert.Equal(t, 0.9, out[0].Confidence)
}

func TestFilterUnsafe(t *testing.T) {
	in := []Suggestion{
		{Text: "git push", Confidence: 0.9},
		{Text: "rm -rf /", Confidence: 0.8},
		{Text: "git reset --hard", Confidence: 0.7},
		{Text: "git push --force", Confidence: 0.6},
	}
	out := filterUnsafe(in, "git commit -m 'wip'")
	require.Len(t, out, 1)
	assert.Equal(t, "git push", out[0].Text)
}

func TestFilterUnsafeAllowsWhenOriginalDestructive(t *testing.T) {
	in := []Suggestion{
		{Text: "rm -rf /tmp/old", Confidence: 0.9},
	}
	out := filterUnsafe(in, "rm -rf /tmp/cache")
	require.Len(t, out, 1)
}

func TestRankByConfidence(t *testing.T) {
	in := []Suggestion{
		{Text: "a", Confidence: 0.3},
		{Text: "b", Confidence: 0.9},
		{Text: "c", Confidence: 0.6},
	}
	out := rank(in)
	assert.Equal(t, "b", out[0].Text)
	assert.Equal(t, "c", out[1].Text)
	assert.Equal(t, "a", out[2].Text)
}

func TestLimitCaps(t *testing.T) {
	in := []Suggestion{
		{Text: "a", Confidence: 0.9, Primary: true},
		{Text: "b", Confidence: 0.7},
		{Text: "c", Confidence: 0.5},
		{Text: "d", Confidence: 0.3},
		{Text: "r", Confidence: 0.8, Recovery: true},
		{Text: "r2", Confidence: 0.6, Recovery: true},
	}
	res := limit(in)
	require.NotNil(t, res.Primary)
	assert.Equal(t, "a", res.Primary.Text)
	assert.Len(t, res.Secondary, 2)
	assert.Equal(t, "b", res.Secondary[0].Text)
	assert.Equal(t, "c", res.Secondary[1].Text)
	require.NotNil(t, res.Recovery)
	assert.Equal(t, "r", res.Recovery.Text)
}

func TestApplyFullPipeline(t *testing.T) {
	in := []Suggestion{
		{Text: "git push", Confidence: 0.9, Primary: true},
		{Text: "git push", Confidence: 0.5},
		{Text: "git log", Confidence: 0.4},
		{Text: "rm -rf /", Confidence: 0.8},
		{Text: "git stash", Confidence: 0.3},
	}
	res := Apply(in, Options{OriginalCmd: "git commit -m 'fix'"})
	require.NotNil(t, res.Primary)
	assert.Equal(t, "git push", res.Primary.Text)
	assert.Len(t, res.Secondary, 2)
	assert.Nil(t, res.Recovery)
}

func TestApplyGenericCaps(t *testing.T) {
	in := []Suggestion{
		{Text: "git push", Confidence: 0.9},
	}
	res := Apply(in, Options{IsGeneric: true, OriginalCmd: "git commit"})
	require.NotNil(t, res.Primary)
	assert.Equal(t, 0.6, res.Primary.Confidence)
}

func TestLimitPromotesFirstToPrimary(t *testing.T) {
	in := []Suggestion{
		{Text: "a", Confidence: 0.9},
		{Text: "b", Confidence: 0.5},
	}
	res := limit(in)
	require.NotNil(t, res.Primary)
	assert.True(t, res.Primary.Primary)
	assert.Equal(t, "a", res.Primary.Text)
	assert.Len(t, res.Secondary, 1)
}
