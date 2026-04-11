// Package shell provides shell hook initialization for tip.
//
// InitCode returns the shell-specific hook code that integrates tip
// with the user's shell. The hooks capture command lines in preexec
// and call tip observe with exit codes in precmd.
package shell

import "strings"

// InitCode returns the hook code for the given shell name.
// Returns empty string for unsupported shells.
func InitCode(sh string) string {
	switch strings.ToLower(sh) {
	case "zsh":
		return zshInit
	case "bash":
		return bashInit
	default:
		return ""
	}
}

const zshInit = `# tip shell integration (zsh)
__tip_preexec() {
  __tip_cmd="$1"
}

__tip_precmd() {
  local exit_code=$?
  [[ -z "$__tip_cmd" ]] && return
  tip enabled &>/dev/null || { unset __tip_cmd; return; }
  tip observe --cmd-line "$__tip_cmd" --exit "$exit_code" 2>/dev/null
  unset __tip_cmd
}

autoload -Uz add-zsh-hook
add-zsh-hook preexec __tip_preexec
add-zsh-hook precmd __tip_precmd
`

const bashInit = `# tip shell integration (bash)
__tip_preexec() {
  [[ "$BASH_COMMAND" == "$PROMPT_COMMAND" ]] && return
  __tip_cmd="$BASH_COMMAND"
}

__tip_precmd() {
  local exit_code=$?
  [[ -z "$__tip_cmd" ]] && return
  tip enabled &>/dev/null || { unset __tip_cmd; return; }
  tip observe --cmd-line "$__tip_cmd" --exit "$exit_code" 2>/dev/null
  unset __tip_cmd
}

trap '__tip_preexec' DEBUG
PROMPT_COMMAND="__tip_precmd${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
`
