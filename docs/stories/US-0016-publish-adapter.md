# US-0016: Publish Adapter for Discovery

**Persona:** [Adapter Author](../personas/contributors/adapter-author.md)

---

## Story

As an adapter author, I want to publish my adapter so that others
can install it via `tip install`.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Adapter publishing

  Scenario: Published adapter installable by bare name
    Given I published hop.top/tip-mytool to a Go module host
    When a user runs "tip install mytool"
    Then tip resolves to hop.top/tip-mytool@latest
    And the binary is installed

  Scenario: Published adapter installable by URI
    Given I published github.com/me/tip-mytool
    When a user runs "tip install github.com/me/tip-mytool"
    Then the binary is installed from that URI

  Scenario: Installed adapter appears in adapter list
    Given tip-mytool is installed
    When the user runs "tip adapter list"
    Then "mytool" appears with version from describe
```

---

## E2E Test

- `e2e/adapter_publish_test.go` — validates install resolution
  and adapter list discovery after installation.
