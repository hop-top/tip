# US-0011: Understand Adapter Protocol from Docs

**Persona:** [Adapter Author](../personas/contributors/adapter-author.md)

---

## Story

As an adapter author, I want clear protocol documentation so that I
can implement an adapter without trial-and-error.

---

## Acceptance Criteria (Gherkin)

```gherkin
Feature: Adapter protocol documentation

  Scenario: Protocol spec exists and is complete
    Given the docs/adapter-protocol.md file exists
    Then it documents the describe subcommand contract
    And it documents the suggest subcommand contract
    And it documents the context envelope JSON schema
    And it documents the response JSON schema
    And it documents timeout and error handling rules

  Scenario: Example request/response pairs included
    Given the protocol documentation exists
    Then it includes at least 2 request/response examples
    And examples cover success and error cases
```

---

## E2E Test

- `e2e/adapter_protocol_doc_test.go` — validates protocol doc
  exists and example JSON is valid against schema.
