// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"

	"github.com/clivern/mut/core"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Mut management server",
	Run: func(_ *cobra.Command, _ []string) {
		err := core.Load(config)
		if err != nil {
			panic(err.Error())
		}

		err = core.SetupLogging()
		if err != nil {
			panic(err.Error())
		}

		err = core.RunServer(Static)
		if err != nil {
			panic(fmt.Sprintf("Server error: %s", err.Error()))
		}
	},
}

// init registers the server subcommand and flags.
func init() {
	serverCmd.Flags().StringVarP(
		&config,
		"config",
		"c",
		"config.prod.yml",
		"Absolute path to config file (required)",
	)
	serverCmd.MarkFlagRequired("config")
	rootCmd.AddCommand(serverCmd)
}
