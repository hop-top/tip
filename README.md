# tip

Transparent shell-integrated CLI observation layer. Watches your
commands via shell hooks and provides contextual next-step
suggestions powered by external adapter binaries.

tip never wraps or proxies execution. It observes after the fact
via shell hooks -- zero performance penalty, no TTY interference.

## Install

```sh
go install hop.top/tip/cmd/tip@latest
```

## Quickstart

Add to your `.zshrc` (or `.bashrc`):

```sh
eval "$(tip init zsh)"
```

Restart your shell. Now every command you run gets observed. Try:

```sh
git add .
git commit -m "feat: first commit"
```

tip shows a dimmed suggestion below your output:

```
-> git push
   git log --oneline -5
```

## Modes

| Mode | Behavior |
|------|----------|
| guided | Full pipeline: suggest + render (default) |
| transparent | Record session, skip rendering |
| quiet | Skip entirely |

Set mode globally:

```sh
tip on --mode transparent
```

Per-tool override:

```sh
tip on git --mode guided
tip off docker
```

## Commands

```
tip init <shell>            # output shell hook code
tip on [<tool>] [--mode M]  # enable (opt: per-tool + mode)
tip off [<tool>]            # disable (opt: per-tool)
tip observe                 # called by shell hook (internal)
tip status                  # mode, adapters, session info
tip adapter list            # discovered adapters
tip adapter info <name>     # adapter describe output
tip install <ref>           # install adapter
tip config                  # open/show config
tip upgrade                 # self-upgrade
tip version                 # version info
```

## Adapters

tip discovers external `tip-<tool>` binaries in `$PATH` and
`~/.local/share/tip/adapters/`.

### Install an adapter

```sh
# Bare name (resolves to hop.top/tip-git)
tip install git

# Full URI
tip install github.com/user/tip-custom

# With version
tip install git@0.1.0
```

### Available adapters

| Adapter | Tool |
|---------|------|
| tip-git | git |
| tip-gh | GitHub CLI |
| tip-docker | Docker |
| tip-wrangler | Cloudflare Wrangler |
| tip-mmdc | Mermaid CLI |

### Writing adapters

See [Adapter Protocol](docs/adapter-protocol.md) for the full
specification including subcommands, schemas, and examples.

## Configuration

File: `~/.config/tip/config.toml`

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

## License

See [LICENSE](LICENSE).
