package status

import (
	"github.com/spf13/cobra"
	"hop.top/tip/toggle"
)

// Cmd returns the `tip status` cobra command.
func Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current tip state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := toggle.NewStore()
			// TODO: wire adapter discovery + session store
			// once those packages land from other tracks.
			Write(cmd.OutOrStdout(), Info{
				Store: s,
			})
			return nil
		},
	}
}
