// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cli

import (
	"github.com/clivern/mut/agent/core"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:          "run",
	SilenceUsage: true,
	Short:        "Collect and push cluster state until stopped",
	RunE: func(_ *cobra.Command, _ []string) error {
		return core.RunAgent()
	},
}

// init registers the run subcommand.
func init() {
	rootCmd.AddCommand(runCmd)
}
