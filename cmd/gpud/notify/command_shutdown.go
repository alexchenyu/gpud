package notify

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/config"
	gpudstate "github.com/leptonai/gpud/pkg/gpud-state"
	"github.com/leptonai/gpud/pkg/log"
	"github.com/leptonai/gpud/pkg/sqlite"
)

var cmdShutdown = &cobra.Command{
	Use:   "shutdown",
	Short: "notify machine shutdown",
	RunE:  cmdShutdownFunc,
}

func cmdShutdownFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting notify shutdown command")

	stateFile, err := config.DefaultStateFile()
	if err != nil {
		return fmt.Errorf("failed to get state file: %w", err)
	}

	dbRW, err := sqlite.Open(stateFile)
	if err != nil {
		return fmt.Errorf("failed to open state file: %w", err)
	}
	defer dbRW.Close()

	dbRO, err := sqlite.Open(stateFile, sqlite.WithReadOnly(true))
	if err != nil {
		return fmt.Errorf("failed to open state file: %w", err)
	}
	defer dbRO.Close()

	rootCtx, rootCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer rootCancel()
	machineID, err := gpudstate.ReadMachineIDWithFallback(rootCtx, dbRW, dbRO)
	if err != nil {
		return err
	}

	endpoint, err := gpudstate.ReadMetadata(rootCtx, dbRO, gpudstate.MetadataKeyEndpoint)
	if err != nil {
		return fmt.Errorf("failed to read endpoint: %w", err)
	}
	if endpoint == "" {
		log.Logger.Warn("endpoint is not set, skipping notification")
		os.Exit(0)
	}

	req := payload{
		ID:   machineID,
		Type: NotificationTypeShutdown,
	}

	return notification(endpoint, req)
}
