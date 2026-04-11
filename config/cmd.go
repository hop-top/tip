// Package config provides the tip config command and configuration types.
package config

import (
	"fmt"
	"io"
	"sort"

	"github.com/spf13/cobra"
	"hop.top/kit/xdg"
)

// Info holds data for config display.
type Info struct {
	Path   string
	Values map[string]string
}

// Write renders config information to w.
func Write(w io.Writer, info Info) {
	fmt.Fprintf(w, "Config file: %s\n", info.Path)

	if len(info.Values) == 0 {
		fmt.Fprintln(w, "\n(no configuration set)")
		return
	}

	fmt.Fprintln(w)
	keys := make([]string, 0, len(info.Values))
	for k := range info.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "  %s = %s\n", k, info.Values[k])
	}
}

// Cmd returns the `tip config` cobra command.
func Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show config file path and current values",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfgDir, err := xdg.ConfigDir("tip")
			if err != nil {
				return fmt.Errorf("resolve config dir: %w", err)
			}
			path := cfgDir + "/config.yaml"

			// Load config values via kit/config
			vals := loadValues(path)

			Write(cmd.OutOrStdout(), Info{
				Path:   path,
				Values: vals,
			})
			return nil
		},
	}
}

// loadValues reads the YAML config file and returns key-value pairs.
// Returns nil if file doesn't exist or can't be parsed.
func loadValues(path string) map[string]string {
	cfg := struct {
		Mode    string            `yaml:"mode"`
		Shell   map[string]string `yaml:"shell"`
		Adapter struct {
			Timeout string   `yaml:"timeout"`
			Path    []string `yaml:"path"`
		} `yaml:"adapters"`
		Tools map[string]struct {
			Enabled bool   `yaml:"enabled"`
			Mode    string `yaml:"mode"`
		} `yaml:"tools"`
	}{}

	err := loadYAML(path, &cfg)
	if err != nil {
		return nil
	}

	vals := make(map[string]string)
	if cfg.Mode != "" {
		vals["mode"] = cfg.Mode
	}
	if cfg.Adapter.Timeout != "" {
		vals["adapter.timeout"] = cfg.Adapter.Timeout
	}
	for name, tool := range cfg.Tools {
		if tool.Mode != "" {
			vals["tools."+name+".mode"] = tool.Mode
		}
	}
	if len(vals) == 0 {
		return nil
	}
	return vals
}
