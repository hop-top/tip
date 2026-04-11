package adapter

import (
	"embed"
	"os"
	"path/filepath"
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

// Scaffold creates a new adapter project directory with template files.
// dir is the parent directory; a tip-<name>/ subdirectory is created.
func Scaffold(dir, name string) (string, error) {
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
