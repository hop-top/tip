# US-0006: Control tip Mode Per Tool

**Persona:** [CLI Power User](../personas/individuals/cli-power-user.md)

---

## Story

As a CLI power user, I want to set tip's mode per tool so that I get
guided suggestions for git but silence for docker during builds.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Per-tool mode control

  Scenario: Enable guided for one tool
    When I run "tip on git --mode guided"
    Then git commands produce guided suggestions
    And other tools use the global default mode

  Scenario: Disable specific tool
    Given tip is enabled globally
    When I run "tip off docker"
    Then docker commands produce no suggestions
    And git commands still produce suggestions

  Scenario: Per-tool override persists in session
    Given I ran "tip on git --mode transparent"
    When I run multiple git commands
    Then all git commands use transparent mode
    And session records commands but shows no output

  Scenario: tip status shows per-tool overrides
    Given I have per-tool overrides active
    When I run "tip status"
    Then output shows global mode and per-tool overrides
```

---

## E2E Test

- `e2e/per_tool_mode_test.go` — validates per-tool enable/disable,
  mode override, and status reporting.
