// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the agent version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("mut-agent %s (commit %s, built %s by %s)\n", Version, Commit, Date, BuiltBy)
	},
}

// init registers the version subcommand.
func init() {
	rootCmd.AddCommand(versionCmd)
}
