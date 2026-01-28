// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"os"
	"time"
)

// LoadConfig reads the agent settings from environment variables.
func LoadConfig() (Config, error) {
	interval := time.Minute
	if raw := os.Getenv("INTERVAL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("invalid INTERVAL [%s]: %w", raw, err)
		}

		interval = parsed
	}

	return Config{
		Server:      os.Getenv("SERVER"),
		WorkspaceId: os.Getenv("WORKSPACE"),
		ClusterId:   os.Getenv("CLUSTER"),
		AccessKey:   os.Getenv("ACCESS_KEY"),
		Interval:    interval,
	}, nil
}
