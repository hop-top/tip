package main

import (
	"github.com/spf13/cobra"

	"hop.top/tip/install"
)

var installCmd = &cobra.Command{
	Use:   "install <ref>",
	Short: "Install an adapter via go install",
	Long: `Install an adapter binary.

Resolution rules:
  tip install github.com/x/tip-y       Full URI -> as-is @latest
  tip install git                       Bare name -> hop.top/tip-git@latest
  tip install git@0.1.0                 With version -> @v0.1.0`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return install.Run(args[0])
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
