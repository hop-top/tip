#!/bin/sh
# Run all tip benchmark suites via ben.
# Prerequisite: setup-repo.sh must be run first.
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SUITE_DIR="$SCRIPT_DIR/../suites"

# Setup
sh "$SCRIPT_DIR/setup-repo.sh" /tmp/ben-tip-repo

echo ""
echo "Running tip token benchmarks..."
echo "================================"
echo ""

for suite in "$SUITE_DIR"/*.yaml; do
  name=$(basename "$suite" .yaml)
  echo "--- $name ---"
  ben run --suite "$suite" 2>&1 || echo "  (ben run failed)"
  echo ""
done

echo "================================"
echo "Done. Use 'ben query --last 10' to review results."
