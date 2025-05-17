// Package compact implements the "compact" command.
package compact

import (
	"github.com/spf13/cobra"

	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/log"
)

// Command returns the cobra command for the "compact" command.
func Command() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:   "compact",
	Short: "compact the GPUd state database to reduce the size in disk (GPUd must be stopped)",
	RunE:  cmdRootFunc,
}

func cmdRootFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting compact command")

	return runCompact()
}
