# Persona: Adapter Author (Contributor)

**Primary Role:** Developer who writes `tip-<tool>` adapter binaries
to teach tip about a specific CLI tool's workflows, errors, and
recovery paths.

---

## Goals

- Understand the adapter protocol quickly from docs + examples.
- Ship a working adapter with tests in < 1 day.
- Get adapter merged (for curated) or distributed (for community).

---

## Interaction Pattern

### Scaffold + implement

```pseudocode
// Clone example adapter as template
git clone hop.top/tip-example
mv tip-example tip-mytool && cd tip-mytool
// Edit: describe, suggest, workflows subcommands
// Test with xrr cassettes for deterministic replay
```

### Protocol contract

```pseudocode
// tip-mytool describe (no stdin)
// → stdout JSON: { name, version, commands, capabilities }

// tip-mytool suggest (stdin: context envelope JSON)
// → stdout JSON: { intent, stage, suggestions[], recovery[] }
```

### E2E validation

```pseudocode
// Record real command interactions
XRR_MODE=record tip-mytool suggest < fixtures/commit-success.json
// Replay in tests
XRR_MODE=replay go test ./...
```

---

## Key Pain Points

- Adapter protocol undocumented or ambiguous → trial-and-error.
- No test harness → hard to verify adapter correctness.
- No example adapters → cold start.

---

## System Leverage

### JSON stdin/stdout protocol

Simple contract; any language works. Go adapters can reuse
`hop.top/tip/sdk` helpers (future).

### xrr cassette testing

Record real interactions once; replay deterministically in CI.

### Example adapters

`tip-git`, `tip-gh`, `tip-docker`, `tip-wrangler`, `tip-mmdc` ship
as reference implementations.

---

## User Stories

- [[US-0011]](../../stories/US-0011-adapter-protocol.md)
  — Understand adapter protocol from docs
- [[US-0012]](../../stories/US-0012-scaffold-adapter.md)
  — Scaffold a new adapter from template
- [[US-0013]](../../stories/US-0013-describe-subcommand.md)
  — Implement describe subcommand
- [[US-0014]](../../stories/US-0014-suggest-subcommand.md)
  — Implement suggest subcommand
- [[US-0015]](../../stories/US-0015-test-adapter.md)
  — Test adapter with xrr cassettes
- [[US-0016]](../../stories/US-0016-publish-adapter.md)
  — Publish adapter for discovery

---

## Success Metrics

- Working adapter from scaffold to passing tests in < 4 hours.
- Protocol spec covers 100% of stdin/stdout contract.
- All curated adapters pass the same test harness.

---

## Collaboration with Other Personas

- **CLI Power User:** adapter author enables richer suggestions for
  power users of that tool.
- **CLI Learner:** curated adapters directly improve the learner's
  error recovery and workflow guidance.
