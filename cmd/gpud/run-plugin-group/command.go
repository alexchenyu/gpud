// Package runplugingroup implements the "run-plugin-group" command.
package runplugingroup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	clientv1 "github.com/leptonai/gpud/client/v1"
	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/config"
	"github.com/leptonai/gpud/pkg/log"
)

// Command returns the cobra command for the "run-plugin-group" command.
func Command() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:     "run-plugin-group",
	Short:   "run all components in a plugin group by tag",
	Example: "gpud run-plugin-group <plugin_group_name>",
	RunE:    cmdRootFunc,
}

var flagServerAddr string

func init() {
	cmdRoot.PersistentFlags().StringVar(&flagServerAddr, "server", fmt.Sprintf("https://localhost:%d", config.DefaultGPUdPort), "GPUd server address")
}

func cmdRootFunc(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("exactly one argument (tag_name) is required")
	}

	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting run-plugin-group command")

	if flagServerAddr == "" {
		flagServerAddr = fmt.Sprintf("https://localhost:%d", config.DefaultGPUdPort)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	tagName := args[0]

	// Trigger the component check by tag
	if err := clientv1.TriggerComponentCheckByTag(ctx, flagServerAddr, tagName); err != nil {
		return fmt.Errorf("failed to trigger component check for tag %s: %w", tagName, err)
	}

	fmt.Printf("Successfully triggered component check for tag: %s\n", tagName)
	return nil
}
