// Package main is the entrypoint for the tip CLI.
package main

import (
	"context"
	"os"

	"hop.top/kit/cli"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	root := cli.New(cli.Config{
		Name:    "tip",
		Version: version,
		Short:   "Transparent shell-integrated CLI observation layer",
	})
	if err := root.Execute(context.Background()); err != nil {
		os.Exit(1)
	}
}
