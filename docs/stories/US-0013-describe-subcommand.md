# US-0013: Implement describe Subcommand

**Persona:** [Adapter Author](../personas/contributors/adapter-author.md)

---

## Story

As an adapter author, I want to implement the describe subcommand so
that tip can discover my adapter's capabilities.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Adapter describe subcommand

  Scenario: Describe returns required fields
    Given I implemented tip-mytool describe
    When I run "tip-mytool describe"
    Then stdout JSON contains "name" string
    And stdout JSON contains "version" semver string
    And stdout JSON contains "commands" string array

  Scenario: Describe returns optional capabilities
    Given my adapter supports workflows
    When I run "tip-mytool describe"
    Then stdout JSON "capabilities" includes "workflows"

  Scenario: Describe requires no stdin
    When I run "tip-mytool describe" with empty stdin
    Then it exits 0 with valid JSON
```

---

## E2E Test

- `e2e/adapter_describe_test.go` — validates describe output
  schema for all curated adapters.
