// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	BuiltBy = "unknown"
)

var rootCmd = &cobra.Command{
	Use:           "mut-agent",
	SilenceErrors: true,
	Short: `🐺  Mut Agent that Pushes Kubernetes Cluster State to Management Server.

If you have any suggestions, bug reports, or annoyances please report
them to our issue tracker at <https://github.com/clivern/mut/issues>`,
}

// Execute runs the agent command tree.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
