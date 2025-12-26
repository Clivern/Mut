// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"

	"github.com/clivern/mut/core"

	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the Mut NATS worker",
	Run: func(_ *cobra.Command, _ []string) {
		err := core.Load(config)
		if err != nil {
			panic(err.Error())
		}

		err = core.SetupLogging()
		if err != nil {
			panic(err.Error())
		}

		err = core.RunWorker()
		if err != nil {
			panic(fmt.Sprintf("Worker error: %s", err.Error()))
		}
	},
}

// init registers the worker subcommand and flags.
func init() {
	workerCmd.Flags().StringVarP(
		&config,
		"config",
		"c",
		"config.prod.yml",
		"Absolute path to config file (required)",
	)
	workerCmd.MarkFlagRequired("config")
	rootCmd.AddCommand(workerCmd)
}
