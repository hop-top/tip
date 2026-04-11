// Package install handles adapter installation via go install.
package install

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DefaultVanity is the default module prefix for bare adapter names.
const DefaultVanity = "hop.top/tip-"

// Resolution holds the resolved URI and version for go install.
type Resolution struct {
	URI     string
	Version string
}

// String returns the full go install target.
func (r Resolution) String() string {
	return r.URI + "@" + r.Version
}

// Resolve converts a user-provided ref into a go install target.
//
// Rules:
//  1. Full URI (contains / or .): go install <uri>@<version>
//  2. Bare name: go install hop.top/tip-<name>@latest
//  3. name@version: go install hop.top/tip-<name>@v<version>
func Resolve(ref string) Resolution {
	// Split name@version if present.
	name, version, hasVersion := strings.Cut(ref, "@")

	// Check if it's a full URI (contains / or .).
	if strings.Contains(name, "/") || strings.Contains(name, ".") {
		v := "latest"
		if hasVersion {
			v = normalizeVersion(version)
		}
		return Resolution{URI: name, Version: v}
	}

	// Bare name or name@version.
	uri := DefaultVanity + name
	v := "latest"
	if hasVersion {
		v = normalizeVersion(version)
	}
	return Resolution{URI: uri, Version: v}
}

// normalizeVersion ensures the version starts with "v" if it looks
// like a semver (digits). "latest" passes through unchanged.
func normalizeVersion(v string) string {
	if v == "latest" || v == "" {
		return "latest"
	}
	if len(v) > 0 && v[0] >= '0' && v[0] <= '9' {
		return "v" + v
	}
	return v
}

// Run resolves the ref and runs `go install`.
func Run(ref string) error {
	r := Resolve(ref)
	target := r.String()

	cmd := exec.Command("go", "install", target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Fprintf(os.Stderr, "installing %s ...\n", target)
	return cmd.Run()
}
