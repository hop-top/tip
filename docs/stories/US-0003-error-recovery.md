# US-0003: Get Recovery Suggestion on Command Failure

**Persona:** [CLI Learner](../personas/individuals/cli-learner.md)

---

## Story

As a CLI learner, I want to see a recovery suggestion when a command
fails so that I can fix the problem without searching for the error.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Error recovery suggestions

  Scenario: Failed git commit suggests staging
    Given tip is enabled in guided mode
    And tip-git adapter is installed
    When I run "git commit -m 'test'"
    And the command exits with code 1
    And stderr contains "nothing added to commit"
    Then tip renders a recovery line with "fix:" prefix
    And the recovery suggestion includes "git add"

  Scenario: Non-zero exit with no adapter falls back
    Given tip is enabled in guided mode
    And no adapter exists for "mycli"
    When I run "mycli deploy"
    And the command exits with code 1
    Then tip renders no recovery suggestion
    And tip renders no regular suggestions

  Scenario: Recovery suggestions use warm color
    Given tip is enabled in guided mode
    And NO_COLOR is not set
    When a recovery suggestion is rendered
    Then the "fix:" prefix uses amber/warm dim styling
```

---

## E2E Test

- `e2e/error_recovery_test.go` — validates recovery rendering for
  known adapter errors and generic fallback behavior.
