# US-0007: Get Session-Aware Workflow Suggestions

**Persona:** [CLI Power User](../personas/individuals/cli-power-user.md)

---

## Story

As a CLI power user, I want tip to remember my recent commands so
that suggestions reflect where I am in a multi-step workflow.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Session-aware suggestions

  Scenario: Workflow stage advances with commands
    Given tip is enabled in guided mode
    And tip-git adapter is installed
    When I run "git add ." then "git commit -m 'feat'"
    Then the suggestion after commit reflects post-commit stage
    And the primary suggestion is "git push"

  Scenario: Session history included in adapter context
    Given I have run 3 commands in this session
    When tip calls the adapter suggest subcommand
    Then the stdin context envelope includes history array
    And history contains the last 3 commands

  Scenario: Session is scoped to cwd
    Given I ran commands in /project-a
    When I cd to /project-b and run a command
    Then the session for /project-b has empty history
```

---

## E2E Test

- `e2e/session_workflow_test.go` — validates session persistence,
  cwd scoping, and history-based suggestion improvement.
