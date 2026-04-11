# Persona: CLI Power User (Senior Developer / DevOps Engineer)

**Primary Role:** Experienced terminal user who knows their tools
well but wants contextual workflow acceleration and adapter
extensibility.

---

## Goals

- Get workflow-aware suggestions that match their current task stage.
- Toggle tip per-tool (`tip off docker`) when focus is needed.
- Author or customize adapters for internal/niche tools.

---

## Interaction Pattern

### Selective activation

```pseudocode
// Enable only for git, disable for everything else
tip on git --mode guided
tip off

// Or guided globally with transparent for docker
tip on --mode guided
tip on docker --mode transparent
```

### Session-aware workflow

```pseudocode
git add .
// → git commit
git commit -m "feat: add auth"
// → git push
git push
// → gh pr create (session knows: add → commit → push → PR)
```

### Adapter authoring

```pseudocode
// Create custom adapter for internal tool
mkdir tip-mycli && cd tip-mycli
// implement describe + suggest subcommands
go build -o tip-mycli .
mv tip-mycli ~/.local/bin/
// tip auto-discovers via $PATH
```

---

## Key Pain Points

- Existing shell plugins (oh-my-zsh, etc.) are static aliases, not
  context-aware.
- Suggestion tools that can't be silenced per-tool are annoying.
- No extension point for internal/proprietary CLIs.

---

## System Leverage

### Per-tool mode control

`tip on git --mode guided` / `tip off docker` — granular control.

### External adapter protocol

JSON stdin/stdout contract; any language can implement `tip-<tool>`.
Power user can extend tip for any binary.

### Session history

Session DB tracks recent commands per cwd; suggestions improve with
workflow context.

---

## User Stories

- [[US-0006]](../../stories/US-0006-per-tool-mode.md)
  — Control tip mode per tool
- [[US-0007]](../../stories/US-0007-session-workflow.md)
  — Get session-aware workflow suggestions
- [[US-0008]](../../stories/US-0008-adapter-author.md)
  — Author a custom adapter
- [[US-0009]](../../stories/US-0009-transparent-mode.md)
  — Run in transparent mode (session recording only)
- [[US-0010]](../../stories/US-0010-status-check.md)
  — Check tip status (mode, adapters, session)

---

## Success Metrics

- Per-tool toggle latency < 50ms (no perceptible delay).
- Custom adapter discovery in < 1s after binary placed in $PATH.
- Session context improves suggestion relevance over 5+ commands.

---

## Collaboration with Other Personas

- **CLI Learner:** power user's workflow patterns inform adapter
  design that helps learners.
- **Adapter Author:** power user IS often the adapter author for
  their team's internal tools.
