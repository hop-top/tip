# Persona: CLI Learner (Newcomer / Junior Developer)

**Primary Role:** Developer early in their CLI journey who knows basic
commands but misses workflow patterns, flags, and recovery paths.

---

## Goals

- Discover next steps after each command without leaving the terminal.
- Recover from errors without copy-pasting from StackOverflow.
- Build muscle memory for common workflows (git, docker, wrangler).

---

## Interaction Pattern

### First use — shell integration

```pseudocode
// Add to .zshrc
eval "$(tip init zsh)"

// Normal workflow — tip observes transparently
git add .
git commit -m "first feature"
// tip renders dimmed suggestion:
//   → git push
//     git log --oneline -5
```

### Error recovery

```pseudocode
git commit -m "oops"
// exit code 1 — nothing staged
// tip renders:
//   ✗ git commit failed
//   fix: git add -A && git commit -m "oops"
```

---

## Key Pain Points

- Man pages are dense; `--help` output is a wall of text.
- No context-aware guidance — every search is a fresh tab.
- Errors show cryptic messages; no suggested fix inline.

---

## System Leverage

### Zero-config observation

`eval "$(tip init zsh)"` — one line, no prefix, no wrapper.
Every command is observed; suggestions appear only when useful.

### Dimmed inline hints

Non-intrusive; skippable. Learner sees them, power user ignores
them.

---

## User Stories

- [[US-0001]](../../stories/US-0001-shell-setup.md)
  — Set up tip shell integration
- [[US-0002]](../../stories/US-0002-first-suggestion.md)
  — See first suggestion after a command
- [[US-0003]](../../stories/US-0003-error-recovery.md)
  — Get recovery suggestion on command failure
- [[US-0004]](../../stories/US-0004-toggle-off.md)
  — Disable tip when it gets noisy
- [[US-0005]](../../stories/US-0005-install-adapter.md)
  — Install a curated adapter for richer hints

---

## Success Metrics

- Time-to-first-suggestion < 2 min from shell setup.
- Error recovery suggestion shown for ≥ 80% of common failures.
- Zero interference with native command output.

---

## Collaboration with Other Personas

- **CLI Power User:** learner graduates to power user as they absorb
  workflow patterns.
- **Adapter Author:** adapters created by authors directly improve
  learner experience.
