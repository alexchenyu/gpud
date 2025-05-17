// "gpud" implements the "gpud" command-line interface.
package main

import (
	"os"

	"github.com/leptonai/gpud/cmd/gpud/common"
	cmdcompact "github.com/leptonai/gpud/cmd/gpud/compact"
	cmdcustomplugins "github.com/leptonai/gpud/cmd/gpud/custom-plugins"
	cmddown "github.com/leptonai/gpud/cmd/gpud/down"
	cmdjoin "github.com/leptonai/gpud/cmd/gpud/join"
	cmdlistplugins "github.com/leptonai/gpud/cmd/gpud/list-plugins"
	cmdlogin "github.com/leptonai/gpud/cmd/gpud/login"
	cmdnotify "github.com/leptonai/gpud/cmd/gpud/notify"
	cmdprivateip "github.com/leptonai/gpud/cmd/gpud/private-ip"
	cmdrelease "github.com/leptonai/gpud/cmd/gpud/release"
	cmdrun "github.com/leptonai/gpud/cmd/gpud/run"
	cmdrunplugingroup "github.com/leptonai/gpud/cmd/gpud/run-plugin-group"
	cmdscan "github.com/leptonai/gpud/cmd/gpud/scan"
	cmdstatus "github.com/leptonai/gpud/cmd/gpud/status"
	cmdup "github.com/leptonai/gpud/cmd/gpud/up"
	cmdupdate "github.com/leptonai/gpud/cmd/gpud/update"
)

func main() {
	common.RootCommand().AddCommand(
		cmdcompact.Command(),
		cmdcustomplugins.Command(),
		cmddown.Command(),
		cmdjoin.Command(),
		cmdlistplugins.Command(),
		cmdlogin.Command(),
		cmdnotify.Command(),
		cmdprivateip.Command(),
		cmdrelease.Command(),
		cmdrun.Command(),
		cmdrunplugingroup.Command(),
		cmdscan.Command(),
		cmdstatus.Command(),
		cmdup.Command(),
		cmdupdate.Command(),
	)
	if err := common.RootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
