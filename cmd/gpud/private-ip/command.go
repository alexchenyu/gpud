// Package privateip implements the "private-ip" command.
package privateip

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/log"
	"github.com/leptonai/gpud/pkg/netutil"
)

// Command returns the cobra command for the "private-ip" command.
func Command() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:   "private-ip",
	Short: "get private ip addresses of the machine (useful for debugging)",
	RunE:  cmdRootFunc,
}

func cmdRootFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting private-ip command")

	ips, err := netutil.GetPrivateIPs(
		netutil.WithPrefixesToSkip(
			"lo",
			"eni",
			"cali",
			"docker",
			"lepton",
			"tailscale",
		),
		netutil.WithSuffixesToSkip(".calico"),
	)
	if err != nil {
		return err
	}

	ips.RenderTable(os.Stdout)

	return nil
}
