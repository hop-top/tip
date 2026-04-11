package shell

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Cmd returns the `tip init <shell>` cobra command.
func Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <shell>",
		Short: "Output shell hook code",
		Long:  "Print the shell integration code for zsh or bash. Usage: eval \"$(tip init zsh)\"",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			code := InitCode(args[0])
			if code == "" {
				return fmt.Errorf("unsupported shell: %s (supported: zsh, bash)", args[0])
			}
			fmt.Fprint(cmd.OutOrStdout(), code)
			return nil
		},
	}
}
