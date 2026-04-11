// Command tip is a transparent shell-integrated CLI observation layer.
package main

import (
	"context"
	"os"

	"hop.top/kit/cli"
	"hop.top/tip/config"
	"hop.top/tip/shell"
	"hop.top/tip/status"
	"hop.top/tip/toggle"
	tipver "hop.top/tip/version"
)

// Build-time ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	root := cli.New(cli.Config{
		Name:    "tip",
		Version: version,
		Short:   "Transparent CLI observation layer",
	})

	root.Cmd.AddCommand(shell.Cmd())
	root.Cmd.AddCommand(toggle.OnCmd())
	root.Cmd.AddCommand(toggle.OffCmd())
	root.Cmd.AddCommand(status.Cmd())
	root.Cmd.AddCommand(config.Cmd())
	root.Cmd.AddCommand(tipver.Cmd(tipver.Info{
		Version: version,
		Commit:  commit,
		Date:    date,
	}))
	root.Cmd.AddCommand(tipver.UpgradeCmd("tip", version))

	if err := root.Execute(context.Background()); err != nil {
		os.Exit(1)
	}
}
