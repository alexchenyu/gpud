// Package listplugins implements the "list-plugins" command.
package listplugins

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	clientv1 "github.com/leptonai/gpud/client/v1"
	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/config"
	"github.com/leptonai/gpud/pkg/log"
)

// Command returns the cobra command for the "custom-plugins" command.
func Command() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:     "list-plugins",
	Aliases: []string{"lp"},
	Short:   "list all registered custom plugins",
	RunE:    cmdRootFunc,
}

var flagServerAddr string

func init() {
	cmdRoot.PersistentFlags().StringVar(&flagServerAddr, "server", fmt.Sprintf("https://localhost:%d", config.DefaultGPUdPort), "GPUd server address")
}

func cmdRootFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting list-plugins command")

	if flagServerAddr == "" {
		flagServerAddr = fmt.Sprintf("https://localhost:%d", config.DefaultGPUdPort)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Get custom plugins
	plugins, err := clientv1.GetCustomPlugins(ctx, flagServerAddr)
	if err != nil {
		return fmt.Errorf("failed to get custom plugins: %w", err)
	}

	// Print plugins
	if len(plugins) == 0 {
		fmt.Println("No custom plugins registered")
		return nil
	}

	fmt.Println("Registered custom plugins:")
	for name, spec := range plugins {
		fmt.Printf("- %s (Type: %s, Run Mode: %s)\n", name, spec.Type, spec.RunMode)
	}

	return nil
}
