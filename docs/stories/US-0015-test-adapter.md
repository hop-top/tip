# US-0015: Test Adapter with xrr Cassettes

**Persona:** [Adapter Author](../personas/contributors/adapter-author.md)

---

## Story

As an adapter author, I want to test my adapter with recorded
cassettes so that tests are deterministic and fast.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Adapter testing with xrr

  Scenario: Record cassette from real interaction
    Given XRR_MODE=record and XRR_CASSETTE_DIR is set
    When I run tip-mytool suggest with real stdin
    Then a cassette YAML file is written to the dir

  Scenario: Replay cassette in test
    Given XRR_MODE=replay and cassettes exist
    When I run "go test ./..."
    Then tests replay from cassettes without real I/O
    And all tests pass

  Scenario: Cassette miss fails the test
    Given XRR_MODE=replay and a cassette is missing
    When the test tries to read the missing cassette
    Then the test fails with exit code 2
```

---

## E2E Test

- `e2e/adapter_xrr_test.go` — validates record/replay lifecycle
  and cassette miss behavior.
