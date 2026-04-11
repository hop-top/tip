// Package mode defines the operational modes for tip.
//
// Guided runs the full pipeline (parse, session, adapter, policy, render).
// Transparent runs everything except rendering.
// Quiet skips the entire pipeline.
package mode

import (
	"fmt"
	"strings"

	"hop.top/kit/config"
	"hop.top/kit/xdg"
)

// Mode represents a tip operational mode.
type Mode int

const (
	// Guided is the default mode: full pipeline including rendering.
	Guided Mode = iota
	// Transparent records commands and adapter responses but skips render.
	Transparent
	// Quiet skips the entire observation pipeline.
	Quiet
)

// String returns the lowercase name of the mode.
func (m Mode) String() string {
	switch m {
	case Guided:
		return "guided"
	case Transparent:
		return "transparent"
	case Quiet:
		return "quiet"
	default:
		return fmt.Sprintf("unknown(%d)", int(m))
	}
}

// Parse converts a string to a Mode. Matching is case-insensitive.
// Returns an error for unrecognised values.
func Parse(s string) (Mode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "guided", "":
		return Guided, nil
	case "transparent":
		return Transparent, nil
	case "quiet":
		return Quiet, nil
	default:
		return Guided, fmt.Errorf("unknown mode %q", s)
	}
}

// Config is the top-level tip configuration loaded from YAML.
type Config struct {
	Mode string `yaml:"mode"`
}

// LoadConfig loads the tip configuration from the XDG config path.
// Missing files are silently ignored; the zero value defaults to Guided.
func LoadConfig() (Config, error) {
	var cfg Config
	cfgDir, err := xdg.ConfigDir("tip")
	if err != nil {
		return cfg, fmt.Errorf("resolve config dir: %w", err)
	}
	err = config.Load(&cfg, config.Options{
		UserConfigPath: cfgDir + "/config.yaml",
	})
	return cfg, err
}

// Resolve returns the Mode from the loaded config, defaulting to Guided.
func (c Config) Resolve() (Mode, error) {
	return Parse(c.Mode)
}
