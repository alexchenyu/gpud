package update

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/log"
	"github.com/leptonai/gpud/version"
)

var cmdCheck = &cobra.Command{
	Use:   "check",
	Short: "check availability of new version gpud",
	RunE:  cmdCheckFunc,
}

func cmdCheckFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting check command")

	ver, err := version.DetectLatestVersion()
	if err != nil {
		fmt.Printf("failed to detect the latest version: %v\n", err)
		return err
	}

	fmt.Printf("latest version: %s\n", ver)
	return nil
}
