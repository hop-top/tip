# US-0002: See First Suggestion After a Command

**Persona:** [CLI Learner](../personas/individuals/cli-learner.md)

---

## Story

As a CLI learner, I want to see a contextual next-step suggestion
after running a command so that I discover the natural workflow
without searching docs.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Post-command suggestions

  Scenario: Successful command shows primary suggestion
    Given tip is enabled in guided mode
    And tip-git adapter is installed
    When I run "git add ."
    And the command exits with code 0
    Then tip renders a dimmed suggestion line below the output
    And the primary suggestion is "git commit"

  Scenario: Generic adapter provides basic suggestion
    Given tip is enabled in guided mode
    And no curated adapter exists for "curl"
    When I run "curl https://example.com"
    And the command exits with code 0
    Then tip renders a suggestion from the generic adapter
    And confidence is capped at 0.6

  Scenario: Suggestions respect NO_COLOR
    Given NO_COLOR env var is set
    When tip renders suggestions
    Then no ANSI escape codes are present in output
```

---

## E2E Test

- `e2e/suggestion_render_test.go` — validates suggestion output
  format, adapter dispatch, and NO_COLOR compliance.
