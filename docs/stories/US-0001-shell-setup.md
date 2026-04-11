# US-0001: Set Up tip Shell Integration

**Persona:** [CLI Learner](../personas/individuals/cli-learner.md)

---

## Story

As a CLI learner, I want to enable tip in my shell with one line so
that I get suggestions without changing how I type commands.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Shell integration setup

  Scenario: zsh init outputs valid hook code
    Given tip binary is installed
    When I run "tip init zsh"
    Then stdout contains a preexec hook function
    And stdout contains a precmd hook function
    And the output is valid zsh syntax

  Scenario: bash init outputs valid hook code
    Given tip binary is installed
    When I run "tip init bash"
    Then stdout contains a DEBUG trap
    And stdout contains PROMPT_COMMAND assignment
    And the output is valid bash syntax

  Scenario: hook activates after eval
    Given I have eval'd "tip init zsh" in my shell
    When I run any command
    Then tip observe is called after command completes
    And the original command output is unmodified
```

---

## E2E Test

- `e2e/shell_setup_test.go` — validates init output syntax and
  hook invocation via xrr-recorded shell sessions.
