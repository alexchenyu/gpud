// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause
// This file is based on https://github.com/tailscale/tailscale/blob/012933635b43ac41c8ff4340213bdae9abd6d059/clientupdate/clientupdate.go

// Package update implements the "update" commands.
package update

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/log"
	pkgupdate "github.com/leptonai/gpud/pkg/update"
	"github.com/leptonai/gpud/version"
)

// Command returns the cobra command for the "update" command.
func Command() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:   "update",
	Short: "update gpud",
	RunE:  cmdRootFunc,
}

var (
	flagURL         string
	flagNextVersion string
)

func init() {
	cmdRoot.PersistentFlags().StringVar(&flagURL, "url", "", "url for getting a package")
	cmdRoot.PersistentFlags().StringVar(&flagNextVersion, "next-version", "", "set the next version to update")

	cmdRoot.AddCommand(cmdCheck)
}

func cmdRootFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting update command")

	if flagNextVersion == "" {
		var err error
		flagNextVersion, err = version.DetectLatestVersion()
		if err != nil {
			fmt.Printf("Failed to fetch latest version: %v\n", err)
			return err
		}
	}

	if flagURL == "" {
		flagURL = version.DefaultURLPrefix
	}

	return pkgupdate.Update(flagNextVersion, flagURL)
}
