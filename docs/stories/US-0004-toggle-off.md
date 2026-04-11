# US-0004: Disable tip When It Gets Noisy

**Persona:** [CLI Learner](../personas/individuals/cli-learner.md)

---

## Story

As a CLI learner, I want to turn off tip temporarily so that I can
focus without suggestion noise during intensive terminal work.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Toggle tip off/on

  Scenario: tip off disables all suggestions
    Given tip is enabled in guided mode
    When I run "tip off"
    Then subsequent commands produce no tip output
    And commands execute normally

  Scenario: tip on re-enables suggestions
    Given tip is disabled
    When I run "tip on"
    Then subsequent commands produce suggestions again

  Scenario: tip off persists for current session only
    Given I ran "tip off"
    When I open a new shell session
    Then tip is enabled per config default
```

---

## E2E Test

- `e2e/toggle_test.go` — validates on/off lifecycle within a
  session and cross-session config persistence.
