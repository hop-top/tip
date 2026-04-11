# US-0008: Author a Custom Adapter

**Persona:** [CLI Power User](../personas/individuals/cli-power-user.md)

---

## Story

As a CLI power user, I want to write a custom adapter for my team's
internal CLI so that my colleagues get contextual suggestions.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Custom adapter authoring

  Scenario: Adapter discovered from $PATH
    Given I built a binary named "tip-mycli"
    And tip-mycli is in $PATH
    When I run "mycli deploy"
    Then tip discovers and calls tip-mycli suggest

  Scenario: Adapter discovered from plugin dir
    Given tip-mycli is in ~/.local/share/tip/adapters/
    When I run "mycli deploy"
    Then tip discovers and calls tip-mycli suggest

  Scenario: Adapter describe returns metadata
    Given tip-mycli implements the describe subcommand
    When I run "tip adapter info mycli"
    Then output shows name, version, commands from describe

  Scenario: Adapter timeout is enforced
    Given tip-mycli takes >500ms to respond
    When tip calls tip-mycli suggest
    Then tip times out and skips suggestions silently
```

---

## E2E Test

- `e2e/adapter_author_test.go` — validates discovery paths,
  protocol compliance, and timeout enforcement.
