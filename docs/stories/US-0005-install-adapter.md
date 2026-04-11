# US-0005: Install a Curated Adapter

**Persona:** [CLI Learner](../personas/individuals/cli-learner.md)

---

## Story

As a CLI learner, I want to install an adapter for git so that I get
richer, tool-specific suggestions instead of generic hints.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Adapter installation

  Scenario: Install by bare name
    When I run "tip install git"
    Then tip runs "go install hop.top/tip-git@latest"
    And tip-git binary appears in $GOBIN or $PATH

  Scenario: Install by full URI
    When I run "tip install github.com/someone/tip-kubectl"
    Then tip runs "go install github.com/someone/tip-kubectl@latest"

  Scenario: Install with version
    When I run "tip install git@0.1.0"
    Then tip runs "go install hop.top/tip-git@v0.1.0"

  Scenario: Adapter list shows installed
    Given tip-git is installed in $PATH
    When I run "tip adapter list"
    Then output includes "git" with its version

  Scenario: Adapter info shows metadata
    Given tip-git is installed
    When I run "tip adapter info git"
    Then output shows the describe JSON fields
```

---

## E2E Test

- `e2e/adapter_install_test.go` — validates install resolution
  (bare, URI, versioned) and adapter list/info commands.
