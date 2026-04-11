package toggle

import (
	"fmt"

	"github.com/spf13/cobra"
)

// OnCmd returns the `tip on [tool] [--mode M]` cobra command.
func OnCmd() *cobra.Command {
	var mode string
	cmd := &cobra.Command{
		Use:   "on [tool]",
		Short: "Enable tip (globally or per-tool)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := NewStore()
			tool := ""
			if len(args) > 0 {
				tool = args[0]
			}
			if err := s.On(tool, mode); err != nil {
				return err
			}
			if tool == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "tip enabled")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "tip enabled for %s\n", tool)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "", "Mode override (guided, transparent, quiet)")
	return cmd
}

// EnabledCmd returns the `tip enabled` cobra command.
// Exits 0 if tip is globally enabled, 1 otherwise.
func EnabledCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enabled",
		Short: "Check if tip is enabled (exit 0=yes, 1=no)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := NewStore()
			if !s.IsEnabled("") {
				cmd.SilenceUsage = true
				cmd.SilenceErrors = true
				return fmt.Errorf("disabled")
			}
			return nil
		},
	}
}

// OffCmd returns the `tip off [tool]` cobra command.
func OffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "off [tool]",
		Short: "Disable tip (globally or per-tool)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := NewStore()
			tool := ""
			if len(args) > 0 {
				tool = args[0]
			}
			if err := s.Off(tool); err != nil {
				return err
			}
			if tool == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "tip disabled")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "tip disabled for %s\n", tool)
			}
			return nil
		},
	}
}
