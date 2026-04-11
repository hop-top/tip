package config_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"hop.top/tip/config"
)

func TestShowsFilePath(t *testing.T) {
	var buf bytes.Buffer
	config.Write(&buf, config.Info{
		Path: "/home/user/.config/tip/config.yaml",
		Values: map[string]string{
			"mode": "guided",
		},
	})
	out := buf.String()
	assert.Contains(t, out, "/home/user/.config/tip/config.yaml")
}

func TestShowsCurrentValues(t *testing.T) {
	var buf bytes.Buffer
	config.Write(&buf, config.Info{
		Path: "/tmp/tip/config.yaml",
		Values: map[string]string{
			"mode":            "guided",
			"adapter.timeout": "500ms",
		},
	})
	out := buf.String()
	assert.Contains(t, out, "mode")
	assert.Contains(t, out, "guided")
	assert.Contains(t, out, "adapter.timeout")
	assert.Contains(t, out, "500ms")
}

func TestEmptyValues(t *testing.T) {
	var buf bytes.Buffer
	config.Write(&buf, config.Info{
		Path:   "/tmp/tip/config.yaml",
		Values: nil,
	})
	out := buf.String()
	assert.Contains(t, out, "/tmp/tip/config.yaml")
	assert.Contains(t, out, "no configuration")
}
