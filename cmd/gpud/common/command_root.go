package common

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/leptonai/gpud/pkg/log"
	"github.com/leptonai/gpud/version"
)

// RootCommand returns the root command for the "gpud" root command.
func RootCommand() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:   "gpud",
	Short: "GPUd tool",
	Example: `
# to quick scan for your machine health status
gpud scan

# to start gpud as a systemd unit
sudo gpud up
`,
	Version: version.Version,
}

func init() {
	cmdRoot.PersistentFlags().StringP("log-level", "l", "info", "set the logging level [debug, info, warn, error, fatal, panic, dpanic]")
	cmdRoot.PersistentFlags().String("log-file", "", "set the log file path (set empty to stdout/stderr)")

	cmdRoot.PersistentFlags().String("ibstat-command", "", "sets the ibstat command (leave empty for default, useful for testing)")
	cmdRoot.PersistentFlags().String("ibstatus-command", "", "sets the ibstatus command (leave empty for default, useful for testing)")
}

// FlagLogLevel returns the log level flag value.
//
// "FlagSet that applies to this command (local and persistent declared here and by all parents)"
// ref. https://pkg.go.dev/github.com/spf13/cobra#Command.Flags
func FlagLogLevel(cmd *cobra.Command) (zap.AtomicLevel, error) {
	logLvl, err := cmd.Flags().GetString("log-level")
	if err != nil {
		return zap.NewAtomicLevelAt(zap.InfoLevel), err
	}
	return log.ParseLogLevel(logLvl)
}

// FlagLogFile returns the log file flag value.
//
// "FlagSet that applies to this command (local and persistent declared here and by all parents)"
// ref. https://pkg.go.dev/github.com/spf13/cobra#Command.Flags
func FlagLogFile(cmd *cobra.Command) (string, error) {
	return cmd.Flags().GetString("log-file")
}

// CreateLoggerFromFlags creates a logger from the flags.
func CreateLoggerFromFlags(cmd *cobra.Command) (*log.LeptonLogger, zap.AtomicLevel, error) {
	logLevel, err := FlagLogLevel(cmd)
	if err != nil {
		return nil, zap.NewAtomicLevelAt(zap.InfoLevel), err
	}
	logFile, err := FlagLogFile(cmd)
	if err != nil {
		return nil, zap.NewAtomicLevelAt(zap.InfoLevel), err
	}
	return log.CreateLogger(logLevel, logFile), logLevel, nil
}

// FlagIbstatCommand returns the ibstat command flag value.
//
// "FlagSet that applies to this command (local and persistent declared here and by all parents)"
// ref. https://pkg.go.dev/github.com/spf13/cobra#Command.Flags
func FlagIbstatCommand(cmd *cobra.Command) (string, error) {
	return cmd.Flags().GetString("ibstat-command")
}

// FlagIbstatusCommand returns the ibstatus command flag value.
//
// "FlagSet that applies to this command (local and persistent declared here and by all parents)"
// ref. https://pkg.go.dev/github.com/spf13/cobra#Command.Flags
func FlagIbstatusCommand(cmd *cobra.Command) (string, error) {
	return cmd.Flags().GetString("ibstatus-command")
}
