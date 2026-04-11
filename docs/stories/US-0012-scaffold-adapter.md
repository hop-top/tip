# US-0012: Scaffold a New Adapter from Template

**Persona:** [Adapter Author](../personas/contributors/adapter-author.md)

---

## Story

As an adapter author, I want to scaffold a new adapter project so
that I have a working skeleton to start from.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Adapter scaffolding

  Scenario: Scaffold creates working Go project
    When I run "tip adapter new mytool"
    Then a directory "tip-mytool" is created
    And it contains go.mod with module name
    And it contains main.go with describe + suggest stubs
    And it contains a test file with xrr fixture setup
    And "go build" succeeds in the directory

  Scenario: Scaffold describe returns valid metadata
    Given I scaffolded "tip-mytool"
    When I run "tip-mytool describe"
    Then stdout is valid JSON matching the describe schema
```

---

## E2E Test

- `e2e/adapter_scaffold_test.go` — validates scaffold output
  builds and passes basic protocol compliance.
