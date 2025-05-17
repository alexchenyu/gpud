// Package down implements the "down" command.
package down

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmdcommon "github.com/leptonai/gpud/cmd/common"
	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/log"
	pkgsystemd "github.com/leptonai/gpud/pkg/systemd"
	pkgupdate "github.com/leptonai/gpud/pkg/update"
)

// Command returns the cobra command for the "down" command.
func Command() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:   "down",
	Short: "stop gpud systemd unit",
	Long: `# to stop the existing gpud systemd unit
sudo gpud down

# to uninstall gpud
sudo rm /usr/sbin/gpud
sudo rm /etc/systemd/system/gpud.service
`,
	RunE: cmdRootFunc,
}

func cmdRootFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting down command")

	bin, err := os.Executable()
	if err != nil {
		return err
	}
	if err := pkgupdate.RequireRoot(); err != nil {
		fmt.Printf("%s %q requires root to stop gpud (if not run by systemd, manually kill the process with 'pidof gpud')\n", cmdcommon.WarningSign, bin)
		os.Exit(1)
	}
	if !pkgsystemd.SystemctlExists() {
		fmt.Printf("%s requires systemd, if not run by systemd, manually kill the process with 'pidof gpud'\n", cmdcommon.WarningSign)
		os.Exit(1)
	}

	active, err := pkgsystemd.IsActive("gpud.service")
	if err != nil {
		fmt.Printf("%s failed to check if gpud is running: %v\n", cmdcommon.WarningSign, err)
		os.Exit(1)
	}
	if !active {
		fmt.Printf("%s gpud is not running (no-op)\n", cmdcommon.CheckMark)
		os.Exit(0)
	}

	if err := pkgupdate.StopSystemdUnit(); err != nil {
		fmt.Printf("%s failed to stop systemd unit 'gpud.service': %v\n", cmdcommon.WarningSign, err)
		os.Exit(1)
	}

	if err := pkgupdate.DisableGPUdSystemdUnit(); err != nil {
		fmt.Printf("%s failed to disable systemd unit 'gpud.service': %v\n", cmdcommon.WarningSign, err)
		os.Exit(1)
	}

	fmt.Printf("%s successfully stopped gpud\n", cmdcommon.CheckMark)
	return nil
}
