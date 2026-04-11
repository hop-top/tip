# Adapter Protocol

tip discovers external adapter binaries (`tip-<tool>`) via PATH
and plugin directory. Adapters follow a git-style extension model
with two required subcommands.

## Discovery Order

1. `$TIP_ADAPTER_PATH` directories (if set)
2. Default plugin dir: `~/.local/share/tip/adapters/`
3. `$PATH` lookup for `tip-<tool>`
4. Fallback: built-in Level 0 generic adapter

## Subcommands

### `tip-<tool> describe`

No stdin. Returns adapter metadata on stdout as JSON.

**Schema:**

```json
{
  "name": "<tool>",
  "version": "<semver>",
  "tool_version_cmd": "<cmd to get tool version>",
  "commands": ["<subcommand>", "..."],
  "capabilities": ["workflows", "error_patterns"]
}
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | yes | Tool name (matches binary suffix) |
| version | string | yes | Adapter semver |
| tool_version_cmd | string | no | Shell cmd to get tool version |
| commands | []string | no | Known subcommands |
| capabilities | []string | no | Feature flags |

### `tip-<tool> suggest`

Receives context envelope on stdin. Returns suggestions on stdout.

**Request (stdin):**

```json
{
  "cmd": "git",
  "args": ["commit", "-m", "fix: typo"],
  "exit_code": 0,
  "stderr": "",
  "cwd": "/home/user/project",
  "history": [
    {"cmd": "git", "args": ["add", "."], "exit_code": 0}
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| cmd | string | Base command name |
| args | []string | Command arguments |
| exit_code | int | Process exit code |
| stderr | string | Captured stderr (if available) |
| cwd | string | Working directory |
| history | []object | Last 10 session commands |

**Response (stdout):**

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
    {
      "text": "git log --oneline -5",
      "confidence": 0.5
    }
  ],
  "recovery": null
}
```

| Field | Type | Description |
|-------|------|-------------|
| intent | string | Detected user intent |
| stage | string | Workflow stage label |
| suggestions | []object | Ranked suggestions |
| suggestions[].text | string | Command to suggest |
| suggestions[].confidence | float | 0.0-1.0 ranking score |
| suggestions[].primary | bool | At most one per response |
| suggestions[].reason | string | Optional explanation |
| recovery | string/null | Recovery command when exit != 0 |

## Protocol Rules

- JSON lines; single object per response
- Adapter MUST respond within 500ms; tip times out + skips
- Exit code 0 = valid response; non-zero = skip silently
- `confidence` float 0.0-1.0; policy layer uses for ranking
- `primary: true` on at most 1 suggestion
- `recovery` only populated when `exit_code != 0`
- `history` capped at last 10 commands from session

## Examples

### Success: post-commit suggestion

**Request:**
```json
{
  "cmd": "git",
  "args": ["commit", "-m", "feat: add login"],
  "exit_code": 0,
  "cwd": "/home/user/app",
  "history": [
    {"cmd": "git", "args": ["add", "."], "exit_code": 0}
  ]
}
```

**Response:**
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
    }
  ],
  "recovery": null
}
```

### Error: build failure recovery

**Request:**
```json
{
  "cmd": "docker",
  "args": ["build", "-t", "myapp", "."],
  "exit_code": 1,
  "stderr": "denied: access forbidden",
  "cwd": "/home/user/app",
  "history": []
}
```

**Response:**
```json
{
  "intent": "build",
  "stage": "error",
  "suggestions": [],
  "recovery": "docker login registry"
}
```

### Error: adapter timeout

If the adapter takes longer than 500ms, tip kills the process
and skips silently. No output is rendered. The session still
records the original command.

### Error: adapter crash

If the adapter exits with non-zero code, tip treats it as
unavailable. No output is rendered. Falls back to generic
adapter if available.
