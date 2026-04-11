package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// loadYAML reads a YAML file into dst. Returns error if file missing
// or malformed.
func loadYAML(path string, dst any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, dst)
}
