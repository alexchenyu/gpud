// Package release implements the "release" commands.
package release

import (
	"github.com/spf13/cobra"
)

// Command returns the cobra command for the "release" command.
func Command() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:   "release",
	Short: "release gpud",
}

func init() {
	cmdRoot.AddCommand(
		cmdGenKey,
		cmdSignKey,
		cmdSignPackage,
		cmdVerifyKeySignature,
		cmdVerifyPackageSignature,
	)
}
