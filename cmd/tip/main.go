//go:build !e2e

package main

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"

	"hop.top/tip/adapter"
	tipconfig "hop.top/tip/config"
	tipctx "hop.top/tip/context"
	"hop.top/tip/install"
	"hop.top/tip/mode"
	"hop.top/tip/observe"
	"hop.top/tip/policy"
	"hop.top/tip/render"
	"hop.top/tip/session"
	"hop.top/tip/shell"
	"hop.top/tip/status"
	"hop.top/tip/toggle"
	"hop.top/tip/version"
)

var (
	// Set via ldflags at build time.
	buildVersion = "0.0.1-alpha.1"
	buildCommit  = "dev"
	buildDate    = "unknown"
)

func main() {
	root := &cobra.Command{
		Use:   "tip",
		Short: "Transparent shell-integrated CLI observation layer",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		shell.Cmd(),
		observeCmd(),
		toggle.OnCmd(),
		toggle.OffCmd(),
		toggle.EnabledCmd(),
		statusCmd(),
		adapterCmd(),
		installCmd(),
		tipconfig.Cmd(),
		version.Cmd(version.Info{
			Version: buildVersion,
			Commit:  buildCommit,
			Date:    buildDate,
		}),
		version.UpgradeCmd("tip", buildVersion),
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// observeCmd is the internal command called by shell hooks.
func observeCmd() *cobra.Command {
	var cmdLine string
	var exitCode int

	cmd := &cobra.Command{
		Use:    "observe",
		Short:  "Observe a command (called by shell hooks)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmdLine == "" {
				return nil
			}

			// Load mode.
			cfg, _ := mode.LoadConfig()
			m, _ := cfg.Resolve()

			// Check per-tool override.
			parsed := tipctx.Parse(cmdLine)
			if parsed.Cmd == "" {
				return nil
			}
			ts := toggle.NewStore()
			if !ts.IsEnabled(parsed.Cmd) {
				return nil
			}
			if tm := ts.Mode(parsed.Cmd); tm != "" {
				if pm, err := mode.Parse(tm); err == nil {
					m = pm
				}
			}

			// Open session store.
			store, err := session.Open()
			if err != nil {
				// Session failure is non-fatal.
				store = nil
			}
			if store != nil {
				defer store.Close()
			}

			// Discover adapter.
			var adpt observe.Adapter
			adapterPath := adapter.Discover(parsed.Cmd)
			if adapterPath != "" {
				adpt = &externalAdapter{path: adapterPath}
			} else {
				adpt = &genericAdapter{g: adapter.NewGeneric()}
			}

			// Build renderer.
			var rend observe.Renderer
			rend = &lipglossRenderer{}

			cwd, _ := os.Getwd()
			return observe.Run(observe.Params{
				CmdLine:  cmdLine,
				ExitCode: exitCode,
				Cwd:      cwd,
				Mode:     m,
				Store:    store,
				Adapter:  adpt,
				Renderer: rend,
			})
		},
	}
	cmd.Flags().StringVar(&cmdLine, "cmd-line", "", "")
	cmd.Flags().IntVar(&exitCode, "exit", 0, "")
	return cmd
}

// statusCmd wraps the status package with live data.
func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show tip status",
		RunE: func(cmd *cobra.Command, args []string) error {
			ts := toggle.NewStore()

			adapterInfos := adapter.List()
			adapterNames := make([]string, len(adapterInfos))
			for i, a := range adapterInfos {
				adapterNames[i] = a.Name
			}

			var cmdCount int
			if store, err := session.Open(); err == nil {
				defer store.Close()
				cwd, _ := os.Getwd()
				if sess, err := store.FindOrCreateSession(
					cmd.Context(), cwd,
				); err == nil {
					cmds, _ := store.History(
						cmd.Context(), sess.ID, 100,
					)
					cmdCount = len(cmds)
				}
			}

			info := status.Info{
				Store:        ts,
				Adapters:     adapterNames,
				CommandCount: cmdCount,
			}
			status.Write(cmd.OutOrStdout(), info)
			return nil
		},
	}
}

// adapterCmd groups adapter subcommands.
func adapterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adapter",
		Short: "Manage adapters",
	}
	cmd.AddCommand(
		adapterListCmd(),
		adapterInfoCmd(),
		adapterNewCmd(),
	)
	return cmd
}

func adapterListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List discovered adapters",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, info := range adapter.List() {
				fmt.Fprintln(cmd.OutOrStdout(), info.Name)
			}
			return nil
		},
	}
}

func adapterInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <name>",
		Short: "Show adapter metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := adapter.Discover(args[0])
			if path == "" {
				return fmt.Errorf("adapter not found: %s", args[0])
			}
			desc, err := adapter.CallDescribe(path)
			if err != nil {
				return fmt.Errorf("describe failed: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"Name:         %s\n"+
					"Version:      %s\n"+
					"Commands:     %v\n"+
					"Capabilities: %v\n",
				desc.Name, desc.Version,
				desc.Commands, desc.Capabilities)
			return nil
		},
	}
}

func adapterNewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new <name>",
		Short: "Scaffold a new adapter project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			path, err := adapter.Scaffold(cwd, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", path)
			return nil
		},
	}
}

// installCmd wires the install package.
func installCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <ref>",
		Short: "Install an adapter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return install.Run(args[0])
		},
	}
}

// --- Bridge types ---

// externalAdapter bridges adapter.Call to observe.Adapter.
type externalAdapter struct {
	path string
}

func (a *externalAdapter) Suggest(
	env tipctx.Envelope,
) ([]policy.Suggestion, bool, error) {
	// Convert context.Envelope to adapter.ContextEnvelope.
	ae := adapter.ContextEnvelope{
		Cmd:      env.Cmd,
		Args:     env.Args,
		ExitCode: env.ExitCode,
		Stderr:   env.Stderr,
		Cwd:      env.Cwd,
	}
	for _, h := range env.History {
		ae.History = append(ae.History, adapter.HistoryEntry{
			Cmd:      h.Cmd,
			Args:     h.Args,
			ExitCode: h.ExitCode,
		})
	}

	resp, err := adapter.Call(a.path, ae)
	if err != nil || resp == nil {
		return nil, false, err
	}

	return adapterToPolicy(resp), false, nil
}

// genericAdapter bridges adapter.Generic to observe.Adapter.
type genericAdapter struct {
	g *adapter.Generic
}

func (a *genericAdapter) Suggest(
	env tipctx.Envelope,
) ([]policy.Suggestion, bool, error) {
	ae := adapter.ContextEnvelope{
		Cmd:      env.Cmd,
		Args:     env.Args,
		ExitCode: env.ExitCode,
		Stderr:   env.Stderr,
		Cwd:      env.Cwd,
	}
	resp := a.g.Suggest(ae)
	if resp == nil {
		return nil, true, nil
	}
	return adapterToPolicy(resp), true, nil
}

func adapterToPolicy(resp *adapter.SuggestResponse) []policy.Suggestion {
	var out []policy.Suggestion
	for _, s := range resp.Suggestions {
		out = append(out, policy.Suggestion{
			Text:       s.Text,
			Confidence: s.Confidence,
			Primary:    s.Primary,
			Reason:     s.Reason,
		})
	}
	if resp.Recovery != nil {
		out = append(out, policy.Suggestion{
			Text:       resp.Recovery.Text,
			Confidence: resp.Recovery.Confidence,
			Reason:     resp.Recovery.Reason,
			Recovery:   true,
		})
	}
	return out
}

// lipglossRenderer bridges render.Format to observe.Renderer.
type lipglossRenderer struct{}

func (r *lipglossRenderer) Render(result policy.Result) error {
	rr := policyToRender(result)
	out := render.Format(rr, render.Options{
		NoColor: os.Getenv("NO_COLOR") != "",
	})
	if out != "" {
		fmt.Print(out)
	}
	return nil
}

func policyToRender(result policy.Result) render.Result {
	var rr render.Result
	if result.Primary != nil {
		rr.Suggestions = append(rr.Suggestions, render.Suggestion{
			Text:    result.Primary.Text,
			Primary: true,
		})
	}
	for _, s := range result.Secondary {
		rr.Suggestions = append(rr.Suggestions, render.Suggestion{
			Text: s.Text,
		})
	}
	if result.Recovery != nil {
		rr.Recovery = &render.Recovery{
			Fix: result.Recovery.Text,
		}
	}
	return rr
}

