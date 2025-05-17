// Package notify implements the "notify" commands.
package notify

import (
	"github.com/spf13/cobra"
)

// Command returns the cobra command for the "notify" command.
func Command() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:     "notify",
	Aliases: []string{"nt"},
	Short:   "notify control plane of state change",
}

func init() {
	cmdRoot.AddCommand(
		cmdStartup,
		cmdShutdown,
	)
}
