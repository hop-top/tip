package install

import (
	"testing"
)

func TestResolveFullURI(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"github.com/x/tip-y", "github.com/x/tip-y@latest"},
		{"github.com/x/tip-y@v1.0.0", "github.com/x/tip-y@v1.0.0"},
		{"github.com/x/tip-y@1.0.0", "github.com/x/tip-y@v1.0.0"},
		{"hop.top/tip-custom", "hop.top/tip-custom@latest"},
		{"hop.top/tip-custom@latest", "hop.top/tip-custom@latest"},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			r := Resolve(tt.ref)
			got := r.String()
			if got != tt.want {
				t.Errorf("Resolve(%q) = %q, want %q", tt.ref, got, tt.want)
			}
		})
	}
}

func TestResolveBareName(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"git", "hop.top/tip-git@latest"},
		{"docker", "hop.top/tip-docker@latest"},
		{"wrangler", "hop.top/tip-wrangler@latest"},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			r := Resolve(tt.ref)
			got := r.String()
			if got != tt.want {
				t.Errorf("Resolve(%q) = %q, want %q", tt.ref, got, tt.want)
			}
		})
	}
}

func TestResolveNameWithVersion(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"git@0.1.0", "hop.top/tip-git@v0.1.0"},
		{"docker@v2.0.0", "hop.top/tip-docker@v2.0.0"},
		{"gh@latest", "hop.top/tip-gh@latest"},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			r := Resolve(tt.ref)
			got := r.String()
			if got != tt.want {
				t.Errorf("Resolve(%q) = %q, want %q", tt.ref, got, tt.want)
			}
		})
	}
}

func TestResolutionFields(t *testing.T) {
	r := Resolve("git@0.1.0")
	if r.URI != "hop.top/tip-git" {
		t.Errorf("URI = %q", r.URI)
	}
	if r.Version != "v0.1.0" {
		t.Errorf("Version = %q", r.Version)
	}
}
