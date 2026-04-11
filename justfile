default:
    @just --list

# Build the tip binary
build:
    go build -o bin/tip ./cmd/tip

# Run all tests
test *ARGS:
    go test {{ARGS}}

# Run tests for a specific package
test-pkg PKG:
    go test ./{{PKG}}/...

# Run linter
lint:
    go vet ./...

# Format code
fmt:
    gofmt -w .

# Run all checks (lint + test)
check: lint test-all

# Run all package tests
test-all:
    go test ./cmd/tip/... ./mode/... ./session/... ./context/... ./policy/... ./observe/...

# Clean build artifacts
clean:
    rm -rf bin/
