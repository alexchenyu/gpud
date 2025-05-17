// Package scan implements the "scan" command.
package scan

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/log"
	"github.com/leptonai/gpud/pkg/scan"
)

// Command returns the cobra command for the "scan" command.
func Command() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:     "scan",
	Aliases: []string{"check", "s"},
	Short:   "quick scans the host for any major issues",
	RunE:    cmdRootFunc,
}

func cmdRootFunc(cmd *cobra.Command, args []string) error {
	logger, logLevel, err := common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}
	log.Logger = logger

	ibstatCommand, err := common.FlagIbstatCommand(cmd)
	if err != nil {
		return err
	}
	ibstatusCommand, err := common.FlagIbstatusCommand(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting scan command")

	opts := []scan.OpOption{
		scan.WithIbstatCommand(ibstatCommand),
		scan.WithIbstatusCommand(ibstatusCommand),
	}
	if logLevel.Level() <= zap.DebugLevel { // e.g., info, warn, error
		opts = append(opts, scan.WithDebug(true))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err = scan.Scan(ctx, opts...); err != nil {
		return err
	}

	return nil
}
