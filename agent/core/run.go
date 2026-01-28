// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// RunAgent collects and pushes cluster state until SIGINT or SIGTERM.
func RunAgent() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := NewAgent(cfg)
	if err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	return a.Run(ctx)
}
