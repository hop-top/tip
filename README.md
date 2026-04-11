# tip

Your agent just ran 40 commands. How many did it get stuck on?

tip is a shell observation layer that feeds contextual hints to
whatever is driving the terminal — you, an AI agent, or both.
One shell hook. Zero wrappers. Suggestions appear after every
command, dimmed and ignorable.

## Why this matters for agent-augmented work

AI agents execute CLI commands on your behalf. When an agent hits
a failed `docker build` or a rejected `git push`, it either
recovers instantly or burns tokens re-discovering what went wrong.

tip gives the agent (and you) the next step inline:

```
$ git commit -m "feat: auth"

→ git push
  git log --oneline -5
```

```
$ docker build -t app .

✗ failed to resolve source

  fix: docker login registry.io
```

The agent reads tip's output. You read tip's output. Same hint,
same terminal, no extra API call.

## How it works

```
command runs natively
       ↓
shell hook fires (precmd)
       ↓
tip observe --cmd-line "..." --exit $?
       ↓
adapter returns suggestions (JSON, 500ms budget)
       ↓
dimmed hint below prompt
```

tip never wraps, proxies, or intercepts. The binary runs first.
tip reacts to the result.

## Get started

```sh
go install hop.top/tip/cmd/tip@latest
```

One line in `.zshrc` (or `.bashrc`):

```sh
eval "$(tip init zsh)"
```

New shell. Done. Works for you and any agent that inherits
the shell environment.

## Agent integration

Agents that spawn shell sessions (Claude Code, Codex, Aider,
custom harnesses) inherit tip automatically if the shell profile
sources `tip init`. No agent-side configuration needed.

**Transparent mode** records the session without rendering — useful
when an agent drives the terminal and you review later:

```sh
tip on --mode transparent
```

**Guided mode** renders hints the agent can parse from stdout:

```sh
tip on --mode guided
```

**Quiet mode** disables everything — for benchmarks or batch jobs
where observation overhead matters:

```sh
tip on --mode quiet
```

## Adapters

Adapters are standalone binaries (`tip-<tool>`) that understand
a specific CLI's workflows, errors, and recovery paths.

```sh
tip install git       # → hop.top/tip-git@latest
tip install gh        # → hop.top/tip-gh@latest
```

Without a curated adapter, tip falls back to parsing `--help`
output with lower confidence.

Write your own in an afternoon — JSON in, JSON out:

→ [Adapter Protocol](docs/adapter-protocol.md)

| Adapter | Tool |
|---------|------|
| tip-git | git |
| tip-gh | GitHub CLI |
| tip-docker | Docker |
| tip-wrangler | Cloudflare Wrangler |
| tip-mmdc | Mermaid CLI |

## Per-tool control

```sh
tip off docker          # silence docker, keep everything else
tip on git --mode guided
tip off                 # global kill switch
tip on                  # re-enable with configured defaults
```

## Commands

```
tip init <shell>            shell hook (zsh, bash)
tip on [<tool>] [--mode M]  enable globally or per-tool
tip off [<tool>]            disable globally or per-tool
tip status                  mode, adapters, session
tip adapter list            discovered adapters
tip adapter info <name>     adapter metadata
tip adapter new <name>      scaffold new adapter
tip install <ref>           install adapter
tip config                  config path + values
tip version                 version
tip upgrade                 self-upgrade
```

## Configuration

`~/.config/tip/config.yaml`

```yaml
mode: guided

shell:
  enabled: true

adapters:
  path:
    - ~/.local/share/tip/adapters
  timeout: 500ms

tools:
  docker:
    enabled: false
  git:
    mode: guided
```

## License

See [LICENSE](LICENSE).
