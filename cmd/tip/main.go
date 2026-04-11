package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: tip <command>")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "init":
		cmdInit()
	case "observe":
		cmdObserve()
	case "on":
		cmdOn()
	case "off":
		cmdOff()
	case "status":
		cmdStatus()
	case "adapter":
		cmdAdapter()
	case "install":
		cmdInstall()
	case "version":
		fmt.Println("tip 0.0.1-alpha.1")
	case "config":
		fmt.Println("config: ~/.config/tip/config.toml")
	case "upgrade":
		fmt.Println("already up to date")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func cmdInit() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: tip init <shell>")
		os.Exit(1)
	}
	switch os.Args[2] {
	case "zsh":
		fmt.Print(zshHook)
	case "bash":
		fmt.Print(bashHook)
	default:
		fmt.Fprintf(os.Stderr, "unsupported shell: %s\n", os.Args[2])
		os.Exit(1)
	}
}

const zshHook = `__tip_preexec() {
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
`

const bashHook = `__tip_precmd() {
  local exit_code=$?
  [[ -z "$__tip_cmd" ]] && return
  [[ "$__tip_enabled" == "0" ]] && return
  tip observe --cmd-line "$__tip_cmd" --exit "$exit_code" 2>/dev/null
  unset __tip_cmd
}

__tip_preexec() {
  __tip_cmd="$BASH_COMMAND"
}

__tip_enabled=1
trap '__tip_preexec' DEBUG
PROMPT_COMMAND="__tip_precmd;${PROMPT_COMMAND}"
`

func cmdObserve() {
	// Check if disabled
	stateFile := tipStateFile()
	if b, err := os.ReadFile(stateFile); err == nil {
		if strings.TrimSpace(string(b)) == "off" {
			return
		}
	}

	// Check mode
	mode := loadMode()
	if mode == "quiet" {
		return
	}

	// Parse flags
	var cmdLine string
	var exitCode string
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--cmd-line":
			if i+1 < len(os.Args) {
				i++
				cmdLine = os.Args[i]
			}
		case "--exit":
			if i+1 < len(os.Args) {
				i++
				exitCode = os.Args[i]
			}
		}
	}
	if cmdLine == "" {
		return
	}

	// Parse command
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return
	}
	cmd := parts[0]
	args := parts[1:]

	// Record to session DB
	recordSession(cmd, args, exitCode)

	// Transparent mode: record but don't render
	if mode == "transparent" {
		return
	}

	// Discover adapter
	adapterBin := discoverAdapter(cmd)

	if adapterBin == "" {
		return
	}

	// Build context envelope
	envelope := map[string]any{
		"cmd":       cmd,
		"args":      args,
		"exit_code": atoi(exitCode),
		"cwd":       mustCwd(),
		"history":   loadHistory(),
	}

	// Call adapter suggest
	resp := callAdapter(adapterBin, envelope)
	if resp == nil {
		return
	}

	// Render
	renderResponse(resp, exitCode)
}

func cmdOn() {
	if len(os.Args) >= 3 {
		tool := os.Args[2]
		mode := "guided"
		for i := 3; i < len(os.Args); i++ {
			if os.Args[i] == "--mode" && i+1 < len(os.Args) {
				i++
				mode = os.Args[i]
			}
		}
		saveToolConfig(tool, true, mode)
		fmt.Printf("tip enabled for %s (mode: %s)\n", tool, mode)
		return
	}
	_ = os.WriteFile(tipStateFile(), []byte("on"), 0644)
	fmt.Println("tip enabled")
}

func cmdOff() {
	if len(os.Args) >= 3 {
		tool := os.Args[2]
		saveToolConfig(tool, false, "")
		fmt.Printf("tip disabled for %s\n", tool)
		return
	}
	_ = os.WriteFile(tipStateFile(), []byte("off"), 0644)
	fmt.Println("tip disabled")
}

func cmdStatus() {
	state := "on"
	if b, err := os.ReadFile(tipStateFile()); err == nil {
		state = strings.TrimSpace(string(b))
	}
	mode := loadMode()
	if mode == "" {
		mode = "guided"
	}
	fmt.Printf("mode: %s\n", mode)
	fmt.Printf("enabled: %s\n", state)

	adapters := discoverAllAdapters()
	fmt.Printf("adapters: %s\n", strings.Join(adapters, ", "))

	overrides := loadToolOverrides()
	if len(overrides) > 0 {
		fmt.Println("tool overrides:")
		for tool, cfg := range overrides {
			fmt.Printf("  %s: enabled=%v mode=%s\n",
				tool, cfg.Enabled, cfg.Mode)
		}
	}

	fmt.Printf("session: %s\n", mustCwd())
}

func cmdAdapter() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: tip adapter <list|info|new>")
		os.Exit(1)
	}
	switch os.Args[2] {
	case "list":
		adapters := discoverAllAdapters()
		for _, a := range adapters {
			fmt.Println(a)
		}
	case "info":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: tip adapter info <name>")
			os.Exit(1)
		}
		fmt.Printf("adapter: %s\n", os.Args[3])
	case "new":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: tip adapter new <name>")
			os.Exit(1)
		}
		fmt.Printf("scaffold: tip-%s\n", os.Args[3])
	default:
		fmt.Fprintf(os.Stderr,
			"unknown adapter command: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func cmdInstall() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: tip install <ref>")
		os.Exit(1)
	}
	ref := os.Args[2]
	resolved := resolveInstallRef(ref)
	fmt.Printf("install: %s\n", resolved)
}

func resolveInstallRef(ref string) string {
	if strings.Contains(ref, "/") {
		return ref
	}
	if parts := strings.SplitN(ref, "@", 2); len(parts) == 2 {
		return fmt.Sprintf(
			"hop.top/tip-%s@v%s", parts[0], parts[1])
	}
	return fmt.Sprintf("hop.top/tip-%s@latest", ref)
}

// --- helpers ---

func tipStateFile() string {
	dir := os.Getenv("TIP_STATE_DIR")
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "tip-state")
	}
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "state")
}

func tipToolConfigFile() string {
	dir := os.Getenv("TIP_STATE_DIR")
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "tip-state")
	}
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "tools.json")
}

func tipSessionFile() string {
	dir := os.Getenv("TIP_STATE_DIR")
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "tip-state")
	}
	cwd := mustCwd()
	h := fmt.Sprintf("%x", hash(cwd))
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "session-"+h+".json")
}

func tipModeFile() string {
	dir := os.Getenv("TIP_STATE_DIR")
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "tip-state")
	}
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "mode")
}

type toolConfig struct {
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode"`
}

func saveToolConfig(tool string, enabled bool, mode string) {
	overrides := loadToolOverrides()
	overrides[tool] = toolConfig{Enabled: enabled, Mode: mode}
	b, _ := json.Marshal(overrides)
	_ = os.WriteFile(tipToolConfigFile(), b, 0644)
}

func loadToolOverrides() map[string]toolConfig {
	m := make(map[string]toolConfig)
	b, err := os.ReadFile(tipToolConfigFile())
	if err != nil {
		return m
	}
	_ = json.Unmarshal(b, &m)
	return m
}

func loadMode() string {
	b, err := os.ReadFile(tipModeFile())
	if err != nil {
		return "guided"
	}
	return strings.TrimSpace(string(b))
}

func mustCwd() string {
	d, _ := os.Getwd()
	return d
}

func hash(s string) uint32 {
	var h uint32
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
}

func atoi(s string) int {
	n := 0
	neg := false
	for i, c := range s {
		if i == 0 && c == '-' {
			neg = true
			continue
		}
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	if neg {
		return -n
	}
	return n
}

type sessionEntry struct {
	Cmd      string   `json:"cmd"`
	Args     []string `json:"args"`
	ExitCode string   `json:"exit_code"`
}

func recordSession(cmd string, args []string, exitCode string) {
	entries := loadSessionEntries()
	entries = append(entries, sessionEntry{
		Cmd: cmd, Args: args, ExitCode: exitCode,
	})
	if len(entries) > 100 {
		entries = entries[len(entries)-100:]
	}
	b, _ := json.Marshal(entries)
	_ = os.WriteFile(tipSessionFile(), b, 0644)
}

func loadSessionEntries() []sessionEntry {
	var entries []sessionEntry
	b, err := os.ReadFile(tipSessionFile())
	if err != nil {
		return entries
	}
	_ = json.Unmarshal(b, &entries)
	return entries
}

func loadHistory() []map[string]any {
	entries := loadSessionEntries()
	start := 0
	if len(entries) > 10 {
		start = len(entries) - 10
	}
	var hist []map[string]any
	for _, e := range entries[start:] {
		hist = append(hist, map[string]any{
			"cmd":       e.Cmd,
			"args":      e.Args,
			"exit_code": atoi(e.ExitCode),
		})
	}
	return hist
}

func discoverAdapter(cmd string) string {
	name := "tip-" + cmd
	if adapterPath := os.Getenv("TIP_ADAPTER_PATH"); adapterPath != "" {
		for _, dir := range filepath.SplitList(adapterPath) {
			p := filepath.Join(dir, name)
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				return p
			}
		}
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		p := filepath.Join(dir, name)
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}
	return ""
}

func discoverAllAdapters() []string {
	seen := make(map[string]bool)
	var adapters []string

	scan := func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "tip-") && !e.IsDir() {
				name := strings.TrimPrefix(e.Name(), "tip-")
				if !seen[name] {
					seen[name] = true
					adapters = append(adapters, name)
				}
			}
		}
	}

	if adapterPath := os.Getenv("TIP_ADAPTER_PATH"); adapterPath != "" {
		for _, dir := range filepath.SplitList(adapterPath) {
			scan(dir)
		}
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		scan(dir)
	}
	return adapters
}

func callAdapter(
	bin string, envelope map[string]any,
) map[string]any {
	input, _ := json.Marshal(envelope)

	cmd := exec.Command(bin, "suggest")
	cmd.Stdin = strings.NewReader(string(input))
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var resp map[string]any
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil
	}
	return resp
}

func renderResponse(resp map[string]any, exitCode string) {
	noColor := os.Getenv("NO_COLOR") != ""

	suggestions, _ := resp["suggestions"].([]any)
	recovery, _ := resp["recovery"].(string)

	for _, s := range suggestions {
		sm, ok := s.(map[string]any)
		if !ok {
			continue
		}
		text, _ := sm["text"].(string)
		primary, _ := sm["primary"].(bool)
		if text == "" {
			continue
		}
		if primary {
			if noColor {
				fmt.Printf("-> %s\n", text)
			} else {
				fmt.Printf("\033[2m->\033[0m %s\n", text)
			}
		} else {
			if noColor {
				fmt.Printf("   %s\n", text)
			} else {
				fmt.Printf("\033[2m   %s\033[0m\n", text)
			}
		}
	}

	if recovery != "" {
		if noColor {
			fmt.Printf("  fix: %s\n", recovery)
		} else {
			fmt.Printf("\033[2;33m  fix: %s\033[0m\n", recovery)
		}
	}
}
