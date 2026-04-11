# US-0009: Run in Transparent Mode

**Persona:** [CLI Power User](../personas/individuals/cli-power-user.md)

---

## Story

As a CLI power user, I want to run tip in transparent mode so that
my session is recorded for future context without any visible output.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Transparent mode

  Scenario: Transparent mode records but shows nothing
    Given tip is enabled in transparent mode
    When I run "git status"
    Then no tip suggestions appear after the command
    And the session database records the command

  Scenario: Switch from guided to transparent
    Given tip is in guided mode
    When I run "tip on --mode transparent"
    Then subsequent commands produce no tip output
    And session continues recording

  Scenario: Session data available after transparent session
    Given I ran 5 commands in transparent mode
    When I switch to guided mode
    Then session history includes the 5 transparent commands
    And suggestions benefit from that history
```

---

## E2E Test

- `e2e/transparent_mode_test.go` — validates no output, session
  recording, and mode switch continuity.
