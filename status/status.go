// Package status renders tip's current state for the user.
package status

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"hop.top/tip/toggle"
)

// Info holds all data needed to render status output.
type Info struct {
	Store        *toggle.Store
	Adapters     []string
	CommandCount int
	LastActivity string
}

// Write renders status information to w.
func Write(w io.Writer, info Info) {
	global, tools := info.Store.State()

	// Global mode
	if global.Enabled {
		mode := global.Mode
		if mode == "" {
			mode = "guided"
		}
		fmt.Fprintf(w, "Status: enabled (%s)\n", mode)
	} else {
		fmt.Fprintln(w, "Status: disabled")
	}

	// Per-tool overrides
	if len(tools) > 0 {
		fmt.Fprintln(w, "\nTool overrides:")
		names := make([]string, 0, len(tools))
		for name := range tools {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			ts := tools[name]
			mode := ts.Mode
			if mode == "" {
				mode = "default"
			}
			state := "on"
			if !ts.Enabled {
				state = "off"
			}
			fmt.Fprintf(w, "  %s: %s (%s)\n", name, state, mode)
		}
	}

	// Adapters
	if len(info.Adapters) > 0 {
		fmt.Fprintf(w, "\nAdapters: %s\n", strings.Join(info.Adapters, ", "))
	} else {
		fmt.Fprintln(w, "\nAdapters: none discovered")
	}

	// Session
	fmt.Fprintf(w, "\nSession: %d commands", info.CommandCount)
	if info.LastActivity != "" {
		fmt.Fprintf(w, ", last activity %s", info.LastActivity)
	}
	fmt.Fprintln(w)
}
