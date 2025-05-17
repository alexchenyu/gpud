// Package customplugins implements the "custom-plugins" command.
package customplugins

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	apiv1 "github.com/leptonai/gpud/api/v1"
	cmdcommon "github.com/leptonai/gpud/cmd/common"
	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/components"
	nvidiacommon "github.com/leptonai/gpud/pkg/config/common"
	customplugins "github.com/leptonai/gpud/pkg/custom-plugins"
	custompluginstestdata "github.com/leptonai/gpud/pkg/custom-plugins/testdata"
	"github.com/leptonai/gpud/pkg/log"
	nvidianvml "github.com/leptonai/gpud/pkg/nvidia-query/nvml"
)

// Command returns the cobra command for the "custom-plugins" command.
func Command() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:     "custom-plugins",
	Aliases: []string{"cs", "plugin", "plugins"},
	Short:   "checks/runs custom plugins",
	RunE:    cmdRootFunc,
}

var (
	flagRun      bool
	flagFailFast bool
)

func init() {
	cmdRoot.PersistentFlags().BoolVarP(&flagRun, "run", "r", false, "run the custom plugins")
	cmdRoot.PersistentFlags().BoolVarP(&flagFailFast, "fail-fast", "f", true, "fail fast, exit immediately if any plugin returns unhealthy state")
}

func cmdRootFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting custom-plugins command")

	var specs customplugins.Specs
	if len(args) == 0 {
		log.Logger.Infow("using example specs")
		specs = custompluginstestdata.ExampleSpecs()
	} else {
		specs, err = customplugins.LoadSpecs(args[0])
		if err != nil {
			return err
		}
	}

	// execute "init" type plugins first
	sort.Slice(specs, func(i, j int) bool {
		// "init" type first
		if specs[i].Type == "init" && specs[j].Type == "init" {
			return i < j
		}
		return specs[i].Type == "init"
	})

	println()
	specs.PrintValidateResults(os.Stdout, cmdcommon.CheckMark, cmdcommon.WarningSign)
	println()

	if verr := specs.Validate(); verr != nil {
		return verr
	}

	if !flagRun {
		log.Logger.Infow("custom plugins are not run, only validating the specs")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	nvmlInstance, err := nvidianvml.New()
	if err != nil {
		return err
	}

	ibstatCommand, err := common.FlagIbstatCommand(cmd)
	if err != nil {
		return err
	}
	ibstatusCommand, err := common.FlagIbstatusCommand(cmd)
	if err != nil {
		return err
	}

	gpudInstance := &components.GPUdInstance{
		RootCtx:      ctx,
		NVMLInstance: nvmlInstance,
		NVIDIAToolOverwrites: nvidiacommon.ToolOverwrites{
			IbstatCommand:   ibstatCommand,
			IbstatusCommand: ibstatusCommand,
		},
	}

	results, err := specs.ExecuteInOrder(gpudInstance, flagFailFast)
	if err != nil {
		return err
	}

	println()
	for _, rs := range results {
		debugger, ok := rs.(components.CheckResultDebugger)
		if ok {
			fmt.Printf("\n### Component %q output\n\n%s\n\n", rs.ComponentName(), debugger.Debug())
		}
	}

	println()
	fmt.Printf("### Results\n\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetRowLine(true)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Component", "Health State", "Summary", "Error", "Run Mode", "Extra Info"})
	for _, rs := range results {
		healthState := cmdcommon.CheckMark + " " + string(apiv1.HealthStateTypeHealthy)
		if rs.HealthStateType() != apiv1.HealthStateTypeHealthy {
			healthState = cmdcommon.WarningSign + " " + string(rs.HealthStateType())
		}

		err := ""
		runMode := ""
		extraInfo := ""

		states := rs.HealthStates()
		if len(states) > 0 {
			err = states[0].Error
			runMode = string(states[0].RunMode)

			b, _ := json.Marshal(states[0].ExtraInfo)
			extraInfo = string(b)
		}

		table.Append([]string{rs.ComponentName(), healthState, rs.Summary(), err, runMode, extraInfo})
	}
	table.Render()
	println()

	return nil
}
