# US-0010: Check tip Status

**Persona:** [CLI Power User](../personas/individuals/cli-power-user.md)

---

## Story

As a CLI power user, I want to check tip's current status so that I
know which mode is active, what adapters are available, and what
session data exists.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Status command

  Scenario: Status shows global mode
    Given tip is configured with mode "guided"
    When I run "tip status"
    Then output shows "mode: guided"

  Scenario: Status shows per-tool overrides
    Given I ran "tip off docker"
    When I run "tip status"
    Then output shows docker is disabled

  Scenario: Status shows discovered adapters
    Given tip-git and tip-docker are in $PATH
    When I run "tip status"
    Then output lists git and docker adapters

  Scenario: Status shows session info
    Given I have a session with 5 commands for current cwd
    When I run "tip status"
    Then output shows session command count and last activity
```

---

## E2E Test

- `e2e/status_test.go` — validates status output for mode,
  per-tool overrides, adapters, and session info.
