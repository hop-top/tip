package adapter

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

// ScaffoldData holds template variables for adapter scaffolding.
type ScaffoldData struct {
	Name   string // e.g. "git"
	Module string // e.g. "hop.top/tip-git"
}

// validName matches lowercase alphanumeric, hyphens, and underscores.
var validName = regexp.MustCompile(`^[a-z0-9_-]+$`)

// Scaffold creates a new adapter project directory with template files.
// dir is the parent directory; a tip-<name>/ subdirectory is created.
func Scaffold(dir, name string) (string, error) {
	if strings.Contains(name, string(os.PathSeparator)) ||
		strings.Contains(name, "..") ||
		!validName.MatchString(name) {
		return "", fmt.Errorf("invalid adapter name %q: must match [a-z0-9_-]", name)
	}

	data := ScaffoldData{
		Name:   name,
		Module: "hop.top/tip-" + name,
	}

	outDir := filepath.Join(dir, "tip-"+name)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}

	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		return "", err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		content, err := templateFS.ReadFile("templates/" + e.Name())
		if err != nil {
			return "", err
		}

		tmpl, err := template.New(e.Name()).Parse(string(content))
		if err != nil {
			return "", err
		}

		// Output filename: strip .tmpl suffix if present.
		outName := strings.TrimSuffix(e.Name(), ".tmpl")
		outPath := filepath.Join(outDir, outName)

		f, err := os.Create(outPath)
		if err != nil {
			return "", err
		}

		if err := tmpl.Execute(f, data); err != nil {
			f.Close()
			return "", err
		}
		f.Close()
	}

	return outDir, nil
}
