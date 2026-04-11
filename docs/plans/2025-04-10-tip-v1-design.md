# tip V1 Design Document

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans
> to implement this plan task-by-task.

**Goal:** Build `tip` (hop.top/tip) — a transparent, shell-integrated
CLI observation layer that provides contextual next-step suggestions
after every command, powered by external adapter binaries.

**Architecture:** Shell hooks (preexec/precmd) observe commands
post-execution. `tip observe` dispatches to external `tip-<tool>`
adapter binaries via JSON stdin/stdout protocol. A built-in generic
adapter handles tools without curated adapters. Suggestions render
as dimmed inline text below command output.

**Tech Stack:** Go 1.26+, hop.top/kit (cli, config, sqlstore, tui,
upgrade), hop.top/xrr (e2e test recording), charmbracelet/lipgloss

**Version:** 0.0.1-alpha.1

**Vanity:** hop.top/tip

---

## 1. Module Structure

```
hop.top/tip
├── cmd/tip/              # main entrypoint
│   └── main.go
├── observe/              # tip observe command logic
│   └── observe.go
├── adapter/              # adapter discovery + protocol
│   ├── discovery.go      # find tip-<tool> in PATH + plugin dir
│   ├── protocol.go       # JSON stdin/stdout contract
│   ├── generic.go        # Level 0 built-in adapter
│   └── cache.go          # adapter describe cache
├── context/              # context engine
│   ├── envelope.go       # build context envelope
│   └── parse.go          # parse raw command line
├── session/              # session persistence
│   ├── store.go          # SQLite via kit/sqlstore
│   └── migrations.go     # schema migrations
├── policy/               # suggestion ranking + filtering
│   └── policy.go         # merge, rank, filter, cap
├── render/               # output renderer
│   └── render.go         # dimmed inline via lipgloss
├── mode/                 # operation modes
│   └── mode.go           # guided / transparent / quiet
├── install/              # tip install command
│   └── install.go        # URI resolution + go install
├── docs/
│   ├── plans/
│   ├── personas/
│   └── stories/
└── e2e/                  # e2e tests (xrr-backed)
```

---

## 2. Transparent Shell Integration

### Activation

```zsh
# Add to .zshrc
eval "$(tip init zsh)"
```

### Shell Hooks (zsh)

```zsh
__tip_preexec() {
  __tip_cmd="$1"
}

__tip_precmd() {
  local exit_code=$?
  [[ -z "$__tip_cmd" ]] && return
  [[ "$__tip_enabled" == "0" ]] && return
  tip observe --cmd-line "$__tip_cmd" --exit "$exit_code" 2>/dev/null
  unset __tip_cmd
}

__tip_enabled=1
autoload -Uz add-zsh-hook
add-zsh-hook preexec __tip_preexec
add-zsh-hook precmd __tip_precmd
```

### Key property

tip never wraps or proxies execution. It observes after the fact
via shell hooks. The binary runs natively — zero performance penalty,
no TTY interference, no broken pipes.

---

## 3. Adapter Protocol

External adapter model — git-style extension binaries.

### Discovery Order

1. `$PATH` lookup for `tip-<binary>`
2. Plugin dir: `~/.local/share/tip/adapters/tip-<binary>`
3. Fallback: built-in Level 0 generic

### Subcommands

#### `tip-<tool> describe`

No stdin. Returns adapter metadata on stdout:

```json
{
  "name": "git",
  "version": "0.0.1",
  "tool_version_cmd": "git --version",
  "commands": ["commit", "push", "pull", "merge", "rebase"],
  "capabilities": ["workflows", "error_patterns"]
}
```

#### `tip-<tool> suggest`

Receives context envelope on stdin:

```json
{
  "cmd": "git",
  "args": ["commit", "-m", "fix: typo"],
  "exit_code": 0,
  "stderr": "",
  "cwd": "/home/user/project",
  "history": [
    { "cmd": "git", "args": ["add", "."], "exit_code": 0 }
  ]
}
```

Returns suggestions on stdout:

```json
{
  "intent": "commit",
  "stage": "post-commit",
  "suggestions": [
    {
      "text": "git push",
      "confidence": 0.9,
      "primary": true,
      "reason": "unpushed commits on tracked branch"
    },
    { "text": "git log --oneline -5", "confidence": 0.5 }
  ],
  "recovery": null
}
```

### Protocol Rules

- JSON lines, single object per response
- Adapter MUST respond within 500ms or tip times out + skips
- Adapter exit code 0 = valid response; non-zero = skip silently
- `confidence` float 0.0-1.0; policy layer uses for ranking
- `primary: true` on at most 1 suggestion
- `recovery` only populated when `exit_code != 0`
- `history` capped at last 10 commands from session

### Curated V1 Adapters (separate repos)

- `hop.top/tip-git`
- `hop.top/tip-gh`
- `hop.top/tip-docker`
- `hop.top/tip-wrangler`
- `hop.top/tip-mmdc`

---

## 4. Execution Flow

```
1. user types: docker build -t myapp .
2. shell preexec hook:
   - captures cmd line in shell var
3. docker runs natively (user sees real-time output)
4. shell precmd hook:
   - captures exit_code=$?
   - calls: tip observe \
       --cmd-line "docker build -t myapp ." \
       --exit 0
5. tip observe:
   a. parse cmd-line -> cmd="docker", args=[build,-t,myapp,.]
   b. load session from SQLite (kit/sqlstore, XDG path)
   c. discover adapter: tip-docker in $PATH?
   d. build context envelope (cmd, args, exit_code, cwd, history)
   e. pipe envelope to: tip-docker suggest (stdin JSON)
   f. receive suggestion response
   g. pass through policy layer (rank, filter, cap at 3)
   h. update session (append command, intent, stage)
   i. render dimmed suggestions below prompt
```

### Mode behavior

- **guided:** full pipeline (steps a-i)
- **transparent:** steps a-h only (no render)
- **quiet:** skip entirely (no adapter call, no session write)

---

## 5. Session Schema (SQLite via kit/sqlstore)

DB: `<xdg-data>/tip/sessions.db`

```sql
CREATE TABLE sessions (
  id          TEXT PRIMARY KEY,
  cwd         TEXT NOT NULL,
  created_at  INTEGER NOT NULL,
  updated_at  INTEGER NOT NULL
);

CREATE TABLE commands (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id  TEXT NOT NULL REFERENCES sessions(id),
  cmd         TEXT NOT NULL,
  args        TEXT NOT NULL,
  exit_code   INTEGER NOT NULL,
  intent      TEXT,
  stage       TEXT,
  adapter     TEXT,
  created_at  INTEGER NOT NULL
);

CREATE TABLE adapter_cache (
  cmd         TEXT PRIMARY KEY,
  describe    TEXT NOT NULL,
  help_tree   TEXT,
  cached_at   INTEGER NOT NULL
);

CREATE INDEX idx_commands_session
  ON commands(session_id, created_at DESC);
CREATE INDEX idx_sessions_cwd ON sessions(cwd);
```

---

## 6. Policy Layer

Input: suggestions from adapter (or generic fallback).

Steps:

1. **Deduplicate** — same text -> keep highest confidence
2. **Cap confidence** — generic adapter capped at 0.6
3. **Filter unsafe** — drop rm -rf, --force, reset --hard
   unless original command was already destructive
4. **Rank** — sort by confidence descending
5. **Limit** — max 1 primary + 2 secondary + 1 recovery
6. **Mode gate** — transparent: record, don't render; quiet: skip

---

## 7. Rendering (lipgloss)

### Success output

```
-> git push                   <- primary: normal, dim arrow
  git log --oneline -5        <- secondary: fully dimmed
  git stash                   <- secondary: fully dimmed
```

### Error output

```
x docker build failed

  fix: docker login registry  <- recovery: dim warm/amber
  try: docker build --no-cache
```

### Rules

- 1 blank line separator after command output
- Primary: arrow prefix, default foreground, dimmed arrow
- Secondary: 2-space prefix, dim foreground
- Recovery: fix:/try: prefix, dim warm color
- No box, no border
- Max 4 lines (1 primary + 2 secondary + 1 recovery)
- Respects NO_COLOR env var

---

## 8. CLI Commands

```
tip init <shell>            # output shell hook code
tip on [<tool>] [--mode M]  # enable (opt: per-tool + mode)
tip off [<tool>]            # disable (opt: per-tool)
tip observe                 # called by shell hook (internal)
tip status                  # mode, adapters, session info
tip adapter list            # discovered adapters
tip adapter info <name>     # adapter describe output
tip adapter new <name>      # scaffold new adapter project
tip install <ref>           # install adapter (see resolution)
tip config                  # open/show config
tip upgrade                 # self-upgrade via kit/upgrade
tip version                 # version info
```

### `tip install` resolution

1. Full URI: `tip install github.com/x/tip-y` -> as-is
2. Bare name: `tip install git` -> `hop.top/tip-git`
3. With version: `tip install git@0.1.0` -> `@v0.1.0`
4. No version -> `@latest`

---

## 9. Configuration

File: `~/.config/tip/config.toml` (via kit/config)

```toml
mode = "guided"

[shell]
enabled = true

[adapters]
path = ["~/.local/share/tip/adapters"]
timeout = "500ms"

[tools.docker]
enabled = false

[tools.git]
mode = "guided"
```

---

## 10. Dependencies

| Package | Purpose |
|---------|---------|
| hop.top/kit/cli | Cobra CLI scaffolding |
| hop.top/kit/config | XDG config loading |
| hop.top/kit/sqlstore | SQLite session store |
| hop.top/kit/tui | Lipgloss rendering |
| hop.top/kit/upgrade | Self-upgrade mechanism |
| hop.top/xrr | E2E test record/replay |
| charmbracelet/lipgloss | Terminal styling |

---

## 11. Personas

- [CLI Learner](../personas/individuals/cli-learner.md)
  — newcomer discovering workflows
- [CLI Power User](../personas/individuals/cli-power-user.md)
  — experienced dev wanting acceleration + extensibility
- [Adapter Author](../personas/contributors/adapter-author.md)
  — contributor building tool-specific adapters

---

## 12. Stories

### CLI Learner

| ID | Title |
|----|-------|
| US-0001 | [Shell setup](../stories/US-0001-shell-setup.md) |
| US-0002 | [First suggestion](../stories/US-0002-first-suggestion.md) |
| US-0003 | [Error recovery](../stories/US-0003-error-recovery.md) |
| US-0004 | [Toggle off](../stories/US-0004-toggle-off.md) |
| US-0005 | [Install adapter](../stories/US-0005-install-adapter.md) |

### CLI Power User

| ID | Title |
|----|-------|
| US-0006 | [Per-tool mode](../stories/US-0006-per-tool-mode.md) |
| US-0007 | [Session workflow](../stories/US-0007-session-workflow.md) |
| US-0008 | [Adapter author](../stories/US-0008-adapter-author.md) |
| US-0009 | [Transparent mode](../stories/US-0009-transparent-mode.md) |
| US-0010 | [Status check](../stories/US-0010-status-check.md) |

### Adapter Author

| ID | Title |
|----|-------|
| US-0011 | [Protocol docs](../stories/US-0011-adapter-protocol.md) |
| US-0012 | [Scaffold](../stories/US-0012-scaffold-adapter.md) |
| US-0013 | [Describe cmd](../stories/US-0013-describe-subcommand.md) |
| US-0014 | [Suggest cmd](../stories/US-0014-suggest-subcommand.md) |
| US-0015 | [xrr testing](../stories/US-0015-test-adapter.md) |
| US-0016 | [Publish](../stories/US-0016-publish-adapter.md) |
