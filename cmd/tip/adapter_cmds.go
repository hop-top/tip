package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"hop.top/tip/adapter"
)

var adapterCmd = &cobra.Command{
	Use:   "adapter",
	Short: "Manage tip adapters",
}

var adapterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List discovered adapters",
	RunE: func(cmd *cobra.Command, args []string) error {
		cache := adapter.NewDescribeCache()
		infos := adapter.ListWithDescribe(cache)

		if len(infos) == 0 {
			fmt.Println("No adapters found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tVERSION\tPATH")
		for _, info := range infos {
			ver := "-"
			if info.Desc != nil {
				ver = info.Desc.Version
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", info.Name, ver, info.Path)
		}
		return w.Flush()
	},
}

var adapterInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show adapter details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		path := adapter.Discover(name)
		if path == "" {
			return fmt.Errorf("adapter %q not found", name)
		}

		desc, err := adapter.CallDescribe(path)
		if err != nil {
			return fmt.Errorf("describe %q: %w", name, err)
		}
		if desc == nil {
			return fmt.Errorf("adapter %q did not return describe info", name)
		}

		fmt.Printf("Name:           %s\n", desc.Name)
		fmt.Printf("Version:        %s\n", desc.Version)
		if desc.ToolVersionCmd != "" {
			fmt.Printf("Tool version:   %s\n", desc.ToolVersionCmd)
		}
		if len(desc.Commands) > 0 {
			fmt.Printf("Commands:       %v\n", desc.Commands)
		}
		if len(desc.Capabilities) > 0 {
			fmt.Printf("Capabilities:   %v\n", desc.Capabilities)
		}
		fmt.Printf("Path:           %s\n", path)
		return nil
	},
}

var adapterNewCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Scaffold a new adapter project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		outDir, err := adapter.Scaffold(cwd, args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Created adapter project: %s\n", outDir)
		return nil
	},
}

func init() {
	adapterCmd.AddCommand(adapterListCmd)
	adapterCmd.AddCommand(adapterInfoCmd)
	adapterCmd.AddCommand(adapterNewCmd)
	rootCmd.AddCommand(adapterCmd)
}
