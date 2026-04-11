# US-0014: Implement suggest Subcommand

**Persona:** [Adapter Author](../personas/contributors/adapter-author.md)

---

## Story

As an adapter author, I want to implement the suggest subcommand so
that tip can request suggestions from my adapter.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Adapter suggest subcommand

  Scenario: Suggest returns suggestions on success
    Given I implemented tip-mytool suggest
    When stdin contains a context envelope with exit_code 0
    Then stdout JSON contains "suggestions" array
    And each suggestion has "text" and "confidence"
    And at most 1 suggestion has "primary: true"

  Scenario: Suggest returns recovery on failure
    When stdin contains a context envelope with exit_code 1
    Then stdout JSON contains "recovery" array
    And each recovery item has "text" and "reason"

  Scenario: Suggest returns intent and stage
    When stdin contains a valid context envelope
    Then stdout JSON contains "intent" string
    And stdout JSON contains "stage" string

  Scenario: Suggest handles unknown command gracefully
    When stdin contains a context envelope for unknown subcommand
    Then stdout JSON contains empty suggestions array
    And exit code is 0
```

---

## E2E Test

- `e2e/adapter_suggest_test.go` — validates suggest protocol
  compliance across success, error, and unknown cases.
