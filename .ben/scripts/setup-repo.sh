#!/bin/sh
# Creates a test git repo for benchmarking.
# Produces realistic git status/log/diff output.
set -e

REPO="${1:-/tmp/ben-tip-repo}"
rm -rf "$REPO"
mkdir -p "$REPO" && cd "$REPO"

git init -q
git config user.email "bench@tip.test"
git config user.name "bench"

# Initial commit with multiple files
for f in main.go go.mod README.md Makefile; do
  cat > "$f" <<EOF
// $f — benchmark fixture
package main
func main() {}
EOF
done
mkdir -p internal
cat > internal/handler.go <<EOF
package internal
func Handle() error { return nil }
EOF

git add -A && git commit -q -m "feat: initial commit"

# Second commit — add more files
cat > internal/store.go <<EOF
package internal
type Store struct{ path string }
func Open(p string) *Store { return &Store{path: p} }
EOF
git add -A && git commit -q -m "feat: add store"

# Third commit — modify existing
echo "// updated" >> main.go
git add -A && git commit -q -m "fix: update main"

# Unstaged change for diff
echo "// pending change" >> internal/handler.go

echo "$REPO ready (3 commits, 1 unstaged change)"
