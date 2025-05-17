package compact

import (
	"context"
	"fmt"
	"time"

	cmdcommon "github.com/leptonai/gpud/cmd/common"
	"github.com/leptonai/gpud/pkg/config"
	"github.com/leptonai/gpud/pkg/log"
	"github.com/leptonai/gpud/pkg/netutil"
	"github.com/leptonai/gpud/pkg/process"
	"github.com/leptonai/gpud/pkg/sqlite"
	"github.com/leptonai/gpud/pkg/systemd"
)

func runCompact() error {
	if systemd.SystemctlExists() {
		active, err := systemd.IsActive("gpud.service")
		if err != nil {
			return err
		}
		if active {
			return fmt.Errorf("gpud is running (must be stopped before running compact)")
		}
	}

	portOpen := netutil.IsPortOpen(config.DefaultGPUdPort)
	if portOpen {
		return fmt.Errorf("gpud is running on port %d (must be stopped before running compact)", config.DefaultGPUdPort)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		proc, err := process.FindProcessByName(ctx, "gpud")
		cancel()
		if err != nil {
			return err
		}
		if proc != nil {
			return fmt.Errorf("gpud process is running on PID %d (must be stopped before running compact)", proc.PID())
		}
	}

	log.Logger.Infow("successfully checked gpud is not running")

	stateFile, err := config.DefaultStateFile()
	if err != nil {
		return fmt.Errorf("failed to get state file: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := sqlite.RunCompact(ctx, stateFile); err != nil {
		return fmt.Errorf("failed to compact state file: %w", err)
	}

	fmt.Printf("%s successfully compacted state file\n", cmdcommon.CheckMark)
	return nil
}
