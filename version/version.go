// Package version provides tip version and upgrade commands.
package version

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"hop.top/kit/upgrade"
)

// Info holds build-time version metadata.
type Info struct {
	Version string
	Commit  string
	Date    string
}

// Write renders version information to w.
func Write(w io.Writer, info Info) {
	fmt.Fprintf(w, "tip %s\n", info.Version)
	fmt.Fprintf(w, "  commit: %s\n", info.Commit)
	fmt.Fprintf(w, "  built:  %s\n", info.Date)
}

// Cmd returns the `tip version` cobra command.
func Cmd(info Info) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			Write(cmd.OutOrStdout(), info)
		},
	}
}

// UpgradeCmd returns the `tip upgrade` cobra command.
func UpgradeCmd(name, currentVersion string) *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade tip to the latest version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			checker := upgrade.New(
				upgrade.WithBinary(name, currentVersion),
				upgrade.WithGitHub("ideacrafterslabs/tip"),
			)
			return upgrade.RunCLI(
				cmd.Context(),
				checker,
				upgrade.CLIOptions{Out: cmd.ErrOrStderr()},
			)
		},
	}
}
