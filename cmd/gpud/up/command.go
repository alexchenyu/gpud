// Package up implements the "up" command.
package up

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	cmdcommon "github.com/leptonai/gpud/cmd/common"
	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/config"
	"github.com/leptonai/gpud/pkg/gpud-manager/systemd"
	gpudstate "github.com/leptonai/gpud/pkg/gpud-state"
	"github.com/leptonai/gpud/pkg/log"
	"github.com/leptonai/gpud/pkg/login"
	pkgmachineinfo "github.com/leptonai/gpud/pkg/machine-info"
	nvidianvml "github.com/leptonai/gpud/pkg/nvidia-query/nvml"
	"github.com/leptonai/gpud/pkg/server"
	"github.com/leptonai/gpud/pkg/sqlite"
	pkdsystemd "github.com/leptonai/gpud/pkg/systemd"
	pkgupdate "github.com/leptonai/gpud/pkg/update"
)

// Command returns the cobra command for the "up" command.
func Command() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:   "up",
	Short: "initialize and start gpud in a daemon mode (systemd)",
	Example: `# to start gpud as a systemd unit (recommended)
sudo gpud up

# to enable machine monitoring powered by lepton.ai platform
# sign up here: https://lepton.ai
sudo gpud up --token <LEPTON_AI_TOKEN>

# to start gpud without a systemd unit (e.g., mac)
gpud run

# or
nohup sudo gpud run &>> <your log file path> &
`,
	RunE: cmdRootFunc,
}

var (
	flagToken     string
	flagEndpoint  string
	flagMachineID string
	flagGPUCount  string
	flagPrivateIP string
	flagPublicIP  string
)

func init() {
	cmdRoot.PersistentFlags().StringVar(&flagToken, "token", "", "lepton.ai workspace token for checking in")
	cmdRoot.PersistentFlags().StringVar(&flagEndpoint, "endpoint", "mothership-machine.app.lepton.ai", "endpoint for control plane")
	cmdRoot.PersistentFlags().StringVar(&flagMachineID, "machine-id", "", "machine ID for checking in (only to override default machine id)")
	cmdRoot.PersistentFlags().StringVar(&flagGPUCount, "gpu-count", "", "number of GPUs")
	cmdRoot.PersistentFlags().StringVar(&flagPrivateIP, "private-ip", "", "private IP address")
	cmdRoot.PersistentFlags().StringVar(&flagPublicIP, "public-ip", "", "public IP address")
}

func cmdRootFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting up command")

	if flagToken != "" {
		// runs the same steps as "gpud login"
		log.Logger.Debugw("running login operations with token", "token", flagToken)

		rootCtx, rootCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer rootCancel()

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

		// in case the table has not been created
		if err := gpudstate.CreateTableMetadata(rootCtx, dbRW); err != nil {
			return fmt.Errorf("failed to create metadata table: %w", err)
		}

		prevMachineID, err := gpudstate.ReadMachineIDWithFallback(rootCtx, dbRW, dbRO)
		if err != nil {
			return err
		}
		if prevMachineID != "" {
			fmt.Printf("machine ID %s already assigned (skipping login)\n", prevMachineID)
			return nil
		}

		nvmlInstance, err := nvidianvml.New()
		if err != nil {
			return fmt.Errorf("failed to create nvml instance: %w", err)
		}
		defer func() {
			if err := nvmlInstance.Shutdown(); err != nil {
				log.Logger.Debugw("failed to shutdown nvml instance", "error", err)
			}
		}()

		// previous/existing machine ID is not found (can be empty)
		// if specified, the control plane will validate the machine ID
		// otherwise, the control plane will assign a new machine ID
		req, err := pkgmachineinfo.CreateLoginRequest(flagToken, nvmlInstance, flagMachineID, flagGPUCount)
		if err != nil {
			return fmt.Errorf("failed to create login request: %w", err)
		}

		if flagPrivateIP != "" { // overwrite if not empty
			req.Network.PrivateIP = flagPrivateIP
		}
		if flagPublicIP != "" { // overwrite if not empty
			req.Network.PublicIP = flagPublicIP
		}

		// machine ID has not been assigned yet
		// thus request one and blocks until the login request is processed
		loginResp, err := login.SendRequest(rootCtx, flagEndpoint, *req)
		if err != nil {
			return err
		}

		// persist only after the successful login
		if err := gpudstate.SetMetadata(rootCtx, dbRW, gpudstate.MetadataKeyEndpoint, flagEndpoint); err != nil {
			return fmt.Errorf("failed to record endpoint: %w", err)
		}
		if err := gpudstate.SetMetadata(rootCtx, dbRW, gpudstate.MetadataKeyMachineID, loginResp.MachineID); err != nil {
			return fmt.Errorf("failed to record machine ID: %w", err)
		}
		if err := gpudstate.SetMetadata(rootCtx, dbRW, gpudstate.MetadataKeyToken, loginResp.Token); err != nil {
			return fmt.Errorf("failed to record session token: %w", err)
		}
		if err := gpudstate.SetMetadata(rootCtx, dbRW, gpudstate.MetadataKeyPublicIP, req.Network.PublicIP); err != nil {
			return fmt.Errorf("failed to record public IP: %w", err)
		}
		if err := gpudstate.SetMetadata(rootCtx, dbRW, gpudstate.MetadataKeyPrivateIP, req.Network.PrivateIP); err != nil {
			return fmt.Errorf("failed to record private IP: %w", err)
		}

		fifoFile, err := config.DefaultFifoFile()
		if err != nil {
			return fmt.Errorf("failed to get fifo file: %w", err)
		}

		// for GPUd >= v0.5, we assume "gpud login" first
		// and then "gpud up"
		// we still need this in case "gpud up" and then "gpud login" afterwards
		if err := server.WriteToken(flagToken, fifoFile); err != nil {
			log.Logger.Debugw("failed to write token -- login before first gpud run/up", "error", err)
		}

		fmt.Printf("%s successfully logged in with machine id %s\n", cmdcommon.CheckMark, loginResp.MachineID)
	}

	bin, err := os.Executable()
	if err != nil {
		return err
	}
	if err := pkgupdate.RequireRoot(); err != nil {
		return err
	}
	if !pkdsystemd.SystemctlExists() {
		return fmt.Errorf("requires systemd, to run without systemd, '%s run'", bin)
	}

	if !systemd.DefaultBinExists() {
		return fmt.Errorf("gpud binary not found at %s (you may run 'cp %s %s' to fix the installation)", systemd.DefaultBinPath, bin, systemd.DefaultBinPath)
	}

	if err := systemdInit(flagEndpoint); err != nil {
		return err
	}

	if err := pkgupdate.EnableGPUdSystemdUnit(); err != nil {
		return err
	}

	if err := pkgupdate.RestartGPUdSystemdUnit(); err != nil {
		return err
	}

	log.Logger.Debugw("successfully started gpud (run 'gpud status' for checking status)")
	return nil
}

func systemdInit(endpoint string) error {
	if err := systemd.CreateDefaultEnvFile(endpoint); err != nil {
		return err
	}
	systemdUnitFileData := systemd.GPUDService
	return os.WriteFile(systemd.DefaultUnitFile, []byte(systemdUnitFileData), 0644)
}
