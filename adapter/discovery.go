package adapter

import (
	"os"
	"os/exec"
	"path/filepath"
)

// PluginDir is the default directory for adapter binaries.
const PluginDir = ".local/share/tip/adapters"

// Discover returns the absolute path to the tip-<cmd> adapter binary.
// Discovery order: 1) $PATH, 2) ~/.local/share/tip/adapters/, 3) empty string.
func Discover(cmd string) string {
	return DiscoverWithHome(cmd, "")
}

// DiscoverWithHome is like Discover but accepts an explicit home dir
// (used for testing). If home is empty, os.UserHomeDir is used.
func DiscoverWithHome(cmd, home string) string {
	name := "tip-" + cmd

	// 1. $PATH lookup
	if p, err := exec.LookPath(name); err == nil {
		return p
	}

	// 2. Plugin dir
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return ""
		}
	}
	pluginPath := filepath.Join(home, PluginDir, name)
	if fi, err := os.Stat(pluginPath); err == nil && !fi.IsDir() {
		return pluginPath
	}

	return ""
}
