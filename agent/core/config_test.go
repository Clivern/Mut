// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitLoadConfigReadsEnv(t *testing.T) {
	t.Setenv("SERVER", "https://mut.test")
	t.Setenv("WORKSPACE", "w-1")
	t.Setenv("CLUSTER", "c-1")
	t.Setenv("ACCESS_KEY", "mut_env")
	t.Setenv("INTERVAL", "5s")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "https://mut.test", cfg.Server)
	assert.Equal(t, "w-1", cfg.WorkspaceId)
	assert.Equal(t, "c-1", cfg.ClusterId)
	assert.Equal(t, "mut_env", cfg.AccessKey)
	assert.Equal(t, 5*time.Second, cfg.Interval)
}

func TestUnitLoadConfigDefaultsInterval(t *testing.T) {
	t.Setenv("INTERVAL", "")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, time.Minute, cfg.Interval)
}
